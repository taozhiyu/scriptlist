package service

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
	repository2 "github.com/scriptscat/scriptweb/internal/domain/safe/repository"
	service4 "github.com/scriptscat/scriptweb/internal/domain/safe/service"
	"github.com/scriptscat/scriptweb/internal/domain/script/entity"
	"github.com/scriptscat/scriptweb/internal/domain/script/repository"
	service2 "github.com/scriptscat/scriptweb/internal/domain/script/service"
	service3 "github.com/scriptscat/scriptweb/internal/domain/statistics/service"
	"github.com/scriptscat/scriptweb/internal/domain/user/service"
	request2 "github.com/scriptscat/scriptweb/internal/http/dto/request"
	respond2 "github.com/scriptscat/scriptweb/internal/http/dto/respond"
	"github.com/scriptscat/scriptweb/internal/pkg/errs"
)

type Script interface {
	GetScript(id int64, version string, withcode bool) (*respond2.ScriptInfo, error)
	GetScriptList(search *repository.SearchList, page request2.Pages) (*respond2.List, error)
	GetUserScript(uid int64, self bool, page request2.Pages) (*respond2.List, error)
	GetScriptCodeList(id int64) ([]*respond2.ScriptCode, error)
	GetLatestScriptCode(id int64, withcode bool) (*respond2.ScriptCode, error)
	GetScriptCodeByVersion(id int64, version string, withcode bool) (*respond2.ScriptCode, error)
	GetCategory() ([]*entity.ScriptCategoryList, error)
	AddScore(uid int64, id int64, score *request2.Score) error
	ScoreList(id int64, page *request2.Pages) (*respond2.List, error)
	UserScore(uid, id int64) (*entity.ScriptScore, error)
	CreateScript(uid int64, req *request2.CreateScript) (*entity.Script, error)
	UpdateScript(uid, id int64, req *request2.UpdateScript) error
	UpdateScriptCode(uid, id int64, req *request2.UpdateScriptCode) error
	SyncScript(uid, id int64) error
	FindSyncPrefix(uid int64, prefix string) ([]*entity.Script, error)
}

type script struct {
	userSvc   service.User
	scriptSvc service2.Script
	scoreSvc  service2.Score
	statisSvc service3.Statistics
	rateSvc   service4.Rate
}

func NewScript(userSvc service.User, scriptSvc service2.Script, scoreSvc service2.Score, statisSvc service3.Statistics, rateSvc service4.Rate) Script {
	return &script{
		userSvc:   userSvc,
		scriptSvc: scriptSvc,
		scoreSvc:  scoreSvc,
		statisSvc: statisSvc,
		rateSvc:   rateSvc,
	}
}

func (s *script) GetScript(id int64, version string, withcode bool) (*respond2.ScriptInfo, error) {
	script, err := s.scriptSvc.Info(id)
	if err != nil {
		return nil, err
	}
	user, err := s.userSvc.UserInfo(script.UserId)
	if err != nil {
		return nil, err
	}
	latest, err := s.GetScriptCodeByVersion(id, version, withcode)
	if err != nil {
		return nil, err
	}

	ret := respond2.ToScriptInfo(user, script, latest)
	s.join(ret.Script)
	return ret, nil
}

func (s *script) GetLatestScriptCode(id int64, withcode bool) (*respond2.ScriptCode, error) {
	codes, err := s.scriptSvc.VersionList(id)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, errs.ErrScriptAudit
	}
	user, err := s.userSvc.UserInfo(codes[0].UserId)
	ret := respond2.ToScriptCode(user, codes[0])
	if withcode {
		ret.Meta = codes[0].Meta
		ret.Code = codes[0].Code
		if d, err := s.scriptSvc.GetCodeDefinition(codes[0].ID); err == nil {
			ret.Definition = d.Definition
		}
	}
	return ret, err
}

func (s *script) GetScriptList(search *repository.SearchList, page request2.Pages) (*respond2.List, error) {
	list, total, err := s.scriptSvc.Search(search, page)
	if err != nil {
		return nil, err
	}
	ret := make([]*respond2.Script, len(list))
	for i, v := range list {
		user, _ := s.userSvc.UserInfo(v.UserId)
		latest, err := s.GetLatestScriptCode(v.ID, false)
		if err != nil {
			glog.Errorf("GetLatestScriptCode: %v", err)
		}
		if latest != nil {
			ret[i] = respond2.ToScript(user, v, latest)
			s.join(ret[i])
		}
	}
	return &respond2.List{
		List:  ret,
		Total: total,
	}, nil
}

func (s *script) GetUserScript(uid int64, self bool, page request2.Pages) (*respond2.List, error) {
	list, total, err := s.scriptSvc.UserScript(uid, self, page)
	if err != nil {
		return nil, err
	}
	ret := make([]*respond2.Script, len(list))
	for i, v := range list {
		user, _ := s.userSvc.UserInfo(v.UserId)
		latest, err := s.GetLatestScriptCode(v.ID, false)
		if err != nil {
			glog.Errorf("GetLatestScriptCode: %v", err)
		}
		if latest != nil {
			ret[i] = respond2.ToScript(user, v, latest)
			s.join(ret[i])
		}
	}
	return &respond2.List{
		List:  ret,
		Total: total,
	}, nil
}

func (s *script) join(script *respond2.Script) {
	// 统计
	script.TotalInstall, _ = s.statisSvc.TotalDownload(script.ID)
	script.TodayInstall, _ = s.statisSvc.TodayDownload(script.ID)
	// 评分
	script.Score, _ = s.scoreSvc.GetAvgScore(script.ID)
	script.ScoreNum, _ = s.scoreSvc.Count(script.ID)
}

func (s *script) GetScriptCodeList(id int64) ([]*respond2.ScriptCode, error) {
	list, err := s.scriptSvc.VersionList(id)
	if err != nil {
		return nil, err
	}
	ret := make([]*respond2.ScriptCode, len(list))
	for i, v := range list {
		user, _ := s.userSvc.UserInfo(v.UserId)
		ret[i] = respond2.ToScriptCode(user, v)
	}
	return ret, nil
}

func (s *script) GetScriptCodeByVersion(id int64, version string, withcode bool) (*respond2.ScriptCode, error) {
	if version == "" {
		return s.GetLatestScriptCode(id, withcode)
	}
	list, err := s.scriptSvc.VersionList(id)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Version == version {
			user, _ := s.userSvc.UserInfo(v.UserId)
			ret := respond2.ToScriptCode(user, v)
			if withcode {
				ret.Code = v.Code
				if d, err := s.scriptSvc.GetCodeDefinition(v.ID); err == nil {
					ret.Definition = d.Definition
				}
			}
			return ret, err
		}
	}
	return nil, errs.ErrScriptCodeIsNil
}

func (s *script) GetCategory() ([]*entity.ScriptCategoryList, error) {
	return s.scriptSvc.GetCategory()
}

func (s *script) AddScore(uid int64, id int64, score *request2.Score) error {
	if _, err := s.scriptSvc.Info(id); err != nil {
		return err
	}
	return s.scoreSvc.AddScore(uid, id, score)
}

func (s *script) ScoreList(id int64, page *request2.Pages) (*respond2.List, error) {
	list, total, err := s.scoreSvc.ScoreList(id, page)
	if err != nil {
		return nil, err
	}
	resp := make([]*respond2.ScriptScore, len(list))
	for i, v := range list {
		user, _ := s.userSvc.UserInfo(v.UserId)
		resp[i] = respond2.ToScriptScore(user, v)
	}

	return &respond2.List{
		List:  resp,
		Total: total,
	}, nil
}

func (s *script) UserScore(uid int64, id int64) (*entity.ScriptScore, error) {
	return s.scoreSvc.UserScore(uid, id)
}

func (s *script) CreateScript(uid int64, req *request2.CreateScript) (*entity.Script, error) {
	var ret *entity.Script
	if err := s.rateSvc.Rate(&repository2.RateUserInfo{Uid: uid}, &repository2.RateRule{
		Name:     "post-script",
		Interval: 60,
	}, func() error {
		script, err := s.scriptSvc.CreateScript(uid, req)
		if err != nil {
			return err
		}
		ret = script
		return nil
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *script) UpdateScript(uid, id int64, req *request2.UpdateScript) error {
	return s.rateSvc.Rate(&repository2.RateUserInfo{Uid: uid}, &repository2.RateRule{
		Name:     "update-script",
		Interval: 5,
	}, func() error {
		if err := s.scriptSvc.UpdateScript(uid, id, req); err != nil {
			return err
		}
		if req.SyncMode == service2.SYNC_MODE_MANUAL {
			return s.SyncScript(uid, id)
		}
		return nil
	})
}

func (s *script) UpdateScriptCode(uid, id int64, req *request2.UpdateScriptCode) error {
	return s.rateSvc.Rate(&repository2.RateUserInfo{Uid: uid}, &repository2.RateRule{
		Name:     "update-script-code",
		Interval: 10,
	}, func() error {
		return s.scriptSvc.CreateScriptCode(uid, id, req)
	})
}

func (s *script) SyncScript(uid, id int64) error {
	script, err := s.scriptSvc.Info(id)
	if err != nil {
		return err
	}
	if script.SyncUrl == "" {
		return errs.NewBadRequestError(1000, "同步链接为空")
	}
	return s.rateSvc.Rate(&repository2.RateUserInfo{Uid: uid}, &repository2.RateRule{
		Name:     "sync-script",
		Interval: 10,
	}, func() error {
		req := &request2.UpdateScriptCode{
			//Name:        script.Name,
			//Description: script.Description,
			Content:    script.Content,
			Definition: "",
			Changelog:  "",
			Public:     script.Public,
			Unwell:     script.Unwell,
		}
		req.Code, err = s.requestSyncUrl(script.SyncUrl)
		if err != nil {
			return err
		}
		if script.ContentUrl != "" {
			req.Content, err = s.requestSyncUrl(script.Content)
			if err != nil {
				return err
			}
		}
		if script.Type == entity.LIBRARY_TYPE && script.DefinitionUrl != "" {
			req.Definition, err = s.requestSyncUrl(script.DefinitionUrl)
			if err != nil {
				return err
			}
		}
		return s.scriptSvc.CreateScriptCode(uid, id, req)
	})
}

func (s *script) requestSyncUrl(syncUrl string) (string, error) {
	c := http.Client{
		Timeout: time.Second * 10,
	}
	u, err := url.Parse(syncUrl)
	if err != nil {
		return "", errs.NewErrScriptSyncNetwork(syncUrl, err)
	}
	resp, err := c.Do(&http.Request{
		Method: http.MethodGet,
		URL:    u,
	})
	if err != nil {
		return "", errs.NewErrScriptSyncNetwork(syncUrl, err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(b), nil
}

func (s *script) FindSyncPrefix(uid int64, prefix string) ([]*entity.Script, error) {
	return s.scriptSvc.FindSyncPrefix(uid, prefix)
}
