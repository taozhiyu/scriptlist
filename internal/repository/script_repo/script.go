package script_repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/codfrm/cago/database/db"
	"github.com/codfrm/cago/database/elasticsearch"
	"github.com/codfrm/cago/pkg/utils/httputils"
	entity "github.com/scriptscat/scriptlist/internal/model/entity/script_entity"
)

type SearchOptions struct {
	Keyword  string
	Type     int
	Sort     string
	UserID   int64
	Self     bool
	Category []int64
	Status   int64
	Domain   string
}

type ScriptRepo interface {
	Find(ctx context.Context, id int64) (*entity.Script, error)
	Create(ctx context.Context, script *entity.Script) error
	Update(ctx context.Context, script *entity.Script) error
	Delete(ctx context.Context, id int64) error

	Search(ctx context.Context, options *SearchOptions, page httputils.PageRequest) ([]*entity.Script, int64, error)
}

var defaultScript ScriptRepo

func Script() ScriptRepo {
	return defaultScript
}

func RegisterScript(i ScriptRepo) {
	defaultScript = i
}

type scriptRepo struct {
}

func NewScriptRepo() ScriptRepo {
	return &scriptRepo{}
}

func (u *scriptRepo) Find(ctx context.Context, id int64) (*entity.Script, error) {
	ret := &entity.Script{ID: id}
	if err := db.Ctx(ctx).First(ret).Error; err != nil {
		if db.RecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return ret, nil
}

func (u *scriptRepo) Create(ctx context.Context, script *entity.Script) error {
	return db.Ctx(ctx).Create(script).Error
}

func (u *scriptRepo) Update(ctx context.Context, script *entity.Script) error {
	return db.Ctx(ctx).Updates(script).Error
}

func (u *scriptRepo) Delete(ctx context.Context, id int64) error {
	return db.Ctx(ctx).Delete(&entity.Script{ID: id}).Error
}

func (u *scriptRepo) Search(ctx context.Context, options *SearchOptions, page httputils.PageRequest) ([]*entity.Script, int64, error) {
	if options.Keyword != "" {
		// 暂时不支持排序等
		return u.SearchByEs(ctx, options.Keyword, page)
	}
	// 无关键字从mysql数据库中获取
	list := make([]*entity.Script, 0)
	scriptTbName := (&entity.Script{}).TableName()
	find := db.Ctx(ctx).Model(&entity.Script{})
	if !options.Self {
		find.Where("public=? and unwell=?", entity.PublicScript, entity.Well)
	}
	switch options.Type {
	case 1: // 用户脚本
		find = find.Where("type=1")
	case 2: // 库
		find = find.Where("type=3")
	case 3: // 后台脚本
		options.Category = append(options.Category, 1)
	case 4: // 定时脚本
		options.Category = append(options.Category, 2)
	}
	if len(options.Category) != 0 {
		tabname := db.Default().NamingStrategy.TableName("script_category")
		find = find.Joins("left join "+tabname+" on "+tabname+".script_id="+scriptTbName+".id").
			Where(tabname+".category_id in ?", options.Category)
	}
	if options.Domain != "" {
		tabname := db.Default().NamingStrategy.TableName("script_domain")
		find = find.Joins("left join "+tabname+" on "+tabname+".script_id="+scriptTbName+".id").
			Where(tabname+".domain=?", options.Domain)
	}

	switch options.Sort {
	case "today_download":
		tabname := db.Default().NamingStrategy.TableName("script_date_statistics")
		find = find.Joins(fmt.Sprintf("left join %s on %s.script_id=%s.id and %s.date=?", tabname, tabname, scriptTbName, tabname), time.Now().Format("2006-01-02")).
			Order(tabname + ".download desc,createtime desc")
	case "total_download":
		tabname := db.Default().NamingStrategy.TableName("script_statistics")
		find = find.Joins(fmt.Sprintf("left join %s on %s.script_id=%s.id", tabname, tabname, scriptTbName)).
			Order(tabname + ".download desc,createtime desc")
	case "score":
		tabname := db.Default().NamingStrategy.TableName("script_statistics")
		find = find.Joins(fmt.Sprintf("left join %s on %s.script_id=%s.id", tabname, tabname, scriptTbName)).
			Order(tabname + ".score desc,createtime desc")
	case "updatetime":
		find = find.Where("updatetime>0").Order("updatetime desc,createtime desc")
	default:
		find = find.Order("createtime desc")
	}

	if options.Status > 0 {
		find = find.Where("status=?", options.Status)
	}
	if options.UserID != 0 {
		find = find.Where("user_id=?", options.UserID)
	}
	var num int64
	if err := find.Count(&num).Error; err != nil {
		return nil, 0, err
	}
	find = find.Select(scriptTbName + ".*")
	find = find.Limit(page.GetLimit()).Offset(page.GetOffset())
	if err := find.Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, num, nil
}

// SearchByEs 通过elasticsearch搜索
func (u *scriptRepo) SearchByEs(ctx context.Context, keyword string, page httputils.PageRequest) ([]*entity.Script, int64, error) {
	script := &entity.ScriptSearch{}
	search := elasticsearch.Ctx(ctx).Search
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": keyword,
			},
		},
		"size": page.Limit,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, 0, err
	}
	resp, err := elasticsearch.Ctx(ctx).Search(
		search.WithIndex(script.CollectionName()),
		search.WithBody(&buf),
		search.WithTrackTotalHits(true),
		search.WithPretty())
	if err != nil {
		return nil, 0, err
	}
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch error: [%s] %s", resp.Status(), respByte)
	}
	m := &elasticsearch.SearchResponse[*entity.Script]{}
	if err := json.Unmarshal(respByte, &m); err != nil {
		return nil, 0, err
	}
	ret := make([]*entity.Script, 0)
	for _, v := range m.Hits.Hits {
		ret = append(ret, v.Source)
	}
	return ret, m.Hits.Total.Value, nil
}
