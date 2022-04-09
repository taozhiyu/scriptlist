package subscriber

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/scriptscat/scriptlist/internal/infrastructure/config"
	"github.com/scriptscat/scriptlist/internal/interfaces/api/dto/request"
	"github.com/scriptscat/scriptlist/internal/interfaces/api/dto/respond"
	"github.com/scriptscat/scriptlist/internal/service/issue/broker"
	service5 "github.com/scriptscat/scriptlist/internal/service/issue/service"
	service2 "github.com/scriptscat/scriptlist/internal/service/notify/service"
	broker2 "github.com/scriptscat/scriptlist/internal/service/script/broker"
	service6 "github.com/scriptscat/scriptlist/internal/service/script/service"
	service3 "github.com/scriptscat/scriptlist/internal/service/user/service"
	"github.com/sirupsen/logrus"
)

type ScriptSubscriber struct {
	notifySvc           service2.Sender
	scriptWatchSvc      service6.ScriptWatch
	scriptIssueWatchSvc service5.ScriptIssueWatch
	scriptIssue         service5.Issue
	scriptSvc           service6.Script
	userSvc             service3.User
}

func NewScriptSubscriber(notifySvc service2.Sender, scriptWatchSvc service6.ScriptWatch,
	scriptIssueWatchSvc service5.ScriptIssueWatch, scriptIssue service5.Issue, scriptSvc service6.Script, userSvc service3.User) *ScriptSubscriber {
	return &ScriptSubscriber{
		notifySvc:           notifySvc,
		scriptWatchSvc:      scriptWatchSvc,
		scriptIssueWatchSvc: scriptIssueWatchSvc,
		scriptIssue:         scriptIssue,
		scriptSvc:           scriptSvc,
		userSvc:             userSvc,
	}
}

func (n *ScriptSubscriber) Subscribe(ctx context.Context) error {

	if _, err := broker2.SubscribeEventScriptCreate(n.NotifyScriptCreate); err != nil {
		return err
	}

	if _, err := broker2.SubscribeEventScriptVersionUpdate(n.NotifyScriptUpdate); err != nil {
		return err
	}

	if _, err := broker.SubscribeScriptIssueCreate(n.NotifyScriptIssueCreate); err != nil {
		return err
	}

	if _, err := broker.SubscribeScriptIssueCommentCreate(n.NotifyScriptIssueCommentCreate); err != nil {
		return err
	}

	return nil
}

// NotifyScriptCreate 脚本创建事件,对关注了脚本作者的用户推送脚本创建
func (n *ScriptSubscriber) NotifyScriptCreate(script int64) error {
	scriptInfo, err := n.scriptSvc.Info(script)
	if err != nil {
		return err
	}
	user, err := n.userSvc.UserInfo(scriptInfo.UserId)
	if err != nil {
		return err
	}
	list, _, err := n.userSvc.FollowerList(scriptInfo.UserId, request.AllPage)
	if err != nil {
		return err
	}

	// 脚本作者自己默认关注自己的脚本
	if err := n.scriptWatchSvc.Watch(script, scriptInfo.UserId, service6.ScriptWatchLevelIssueComment); err != nil {
		logrus.Errorf("watch err:%v", err)
	}
	return nil
	// 暂时先屏蔽用户关注
	title := user.Username + " 发布了一个新脚本: " + scriptInfo.Name
	content := fmt.Sprintf("<h2><a href=\"%s\">%s</a></h2><hr/>您可以在<a href='%s'>个人设置页面</a>中取消本邮件的通知,或者取消对该作者的关注",
		config.AppConfig.FrontendUrl+"script-show-page/"+strconv.FormatInt(scriptInfo.ID, 10), scriptInfo.Name,
		config.AppConfig.FrontendUrl+"users/notify",
	)
	for _, v := range list {
		u, err := n.userSvc.SelfInfo(v.Uid)
		if err != nil {
			continue
		}
		uc, err := n.userSvc.GetUserConfig(v.Uid)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyCreateScript].(bool); ok && !n {
			continue
		}
		_ = n.notifySvc.NotifyEmailFrom(user.Username, u.Email, title, content, "text/html")
	}
	return nil
}

// NotifyScriptUpdate 脚本更新事件,对订阅了脚本的进行通知推送
func (n *ScriptSubscriber) NotifyScriptUpdate(script, code int64) error {
	scriptInfo, err := n.scriptSvc.Info(script)
	if err != nil {
		return err
	}
	codeInfo, err := n.scriptSvc.GetCode(code)
	if err != nil {
		return err
	}
	list, err := n.scriptWatchSvc.WatchList(script)
	if err != nil {
		return err
	}
	user, err := n.userSvc.UserInfo(scriptInfo.UserId)
	if err != nil {
		return err
	}

	title := "[" + scriptInfo.Name + "]有新的版本: " + codeInfo.Version
	content := fmt.Sprintf("%s升级到了:%s<hr/><h3>更新日志</h3>%s<hr/><a href=\"%s\">点击查看脚本页面</a><hr/>您可以在<a href='%s'>个人设置页面</a>中取消本邮件的通知,或者取消对该脚本的关注",
		scriptInfo.Name, codeInfo.Version, codeInfo.Changelog,
		config.AppConfig.FrontendUrl+"script-show-page/"+strconv.FormatInt(scriptInfo.ID, 10),
		config.AppConfig.FrontendUrl+"users/notify",
	)
	for uid, v := range list {
		if v < service6.ScriptWatchLevelVersion {
			continue
		}
		u, err := n.userSvc.SelfInfo(uid)
		if err != nil {
			continue
		}
		uc, err := n.userSvc.GetUserConfig(u.UID)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyScriptUpdate].(bool); ok && !n {
			continue
		}
		_ = n.notifySvc.NotifyEmailFrom(user.Username, u.Email, title, content, "text/html")
	}
	return nil
}

// NotifyScriptIssueCreate 脚本反馈创建,对订阅了脚本等级大于等于issue的进行推送,大于等于issueComment的进行反馈关注
func (n *ScriptSubscriber) NotifyScriptIssueCreate(script, issue int64) error {
	scriptInfo, err := n.scriptSvc.Info(script)
	if err != nil {
		return err
	}
	issueInfo, err := n.scriptIssue.GetIssue(issue)
	if err != nil {
		return err
	}
	list, err := n.scriptWatchSvc.WatchList(script)
	if err != nil {
		return err
	}
	user, err := n.userSvc.UserInfo(issueInfo.UserID)
	if err != nil {
		return err
	}
	// issue的创建者监听issue
	if err := n.scriptIssueWatchSvc.Watch(issue, issueInfo.UserID); err != nil {
		logrus.Errorf("issue watch: %v", err)
	}
	title := "[" + scriptInfo.Name + "]" + issueInfo.Title
	content := fmt.Sprintf("%s<hr/><a href=\"%s\">点击查看原文</a><hr/>您可以在<a href='%s'>个人设置页面</a>中取消本邮件的通知,或者取消对该脚本反馈的关注",
		issueInfo.Content,
		config.AppConfig.FrontendUrl+"script-show-page/"+strconv.FormatInt(scriptInfo.ID, 10)+"/issue/"+strconv.FormatInt(issueInfo.ID, 10)+"/comment",
		config.AppConfig.FrontendUrl+"users/notify",
	)
	for uid, level := range list {
		if level < service6.ScriptWatchLevelIssue {
			continue
		}
		// 对issueComment级别的默认监听issue
		if level >= service6.ScriptWatchLevelIssueComment {
			_ = n.scriptIssueWatchSvc.Watch(issue, uid)
		}
		u, err := n.userSvc.SelfInfo(uid)
		if err != nil {
			continue
		}
		uc, err := n.userSvc.GetUserConfig(u.UID)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyScriptIssue].(bool); ok && !n {
			continue
		}
		if err := n.notifySvc.NotifyEmailFrom(user.Username, u.Email, title, content, "text/html"); err != nil {
			logrus.Errorf("sendemail: %v", err)
		}
	}
	// 解析是否有艾特并通知
	users, err := n.parseContent(issueInfo.Content)
	if err != nil {
		logrus.Errorf("parseContent: %v", err)
		return nil
	}
	for _, v := range users {
		uc, err := n.userSvc.GetUserConfig(v.UID)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyAt].(bool); ok && !n {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyScriptIssue].(bool); ok && !n {
			continue
		}
		_ = n.notifySvc.NotifyEmailFrom(user.Username, v.Email, user.Username+" 在 "+issueInfo.Title+" 中有提及到您", content, "text/html")
	}
	return nil
}

// NotifyScriptIssueCommentCreate 脚本反馈评论推送
func (n *ScriptSubscriber) NotifyScriptIssueCommentCreate(issue, comment int64) error {
	commentInfo, err := n.scriptIssue.GetComment(comment)
	if err != nil {
		return err
	}
	issueInfo, err := n.scriptIssue.GetIssue(issue)
	if err != nil {
		return err
	}
	scriptInfo, err := n.scriptSvc.Info(issueInfo.ScriptID)
	if err != nil {
		return err
	}
	list, err := n.scriptIssueWatchSvc.WatchList(issue)
	if err != nil {
		return err
	}
	user, err := n.userSvc.UserInfo(commentInfo.UserID)
	if err != nil {
		return err
	}
	title := "[" + scriptInfo.Name + "]" + issueInfo.Title
	content := fmt.Sprintf("<a href=\"%s\">点击查看原文</a><hr/>您可以在<a href='%s'>个人设置页面</a>中取消本邮件的通知,或者取消对该脚本反馈评论的关注",
		config.AppConfig.FrontendUrl+"script-show-page/"+strconv.FormatInt(scriptInfo.ID, 10)+"/issue/"+strconv.FormatInt(issueInfo.ID, 10)+"/comment",
		config.AppConfig.FrontendUrl+"users/notify",
	)
	switch commentInfo.Type {
	case service5.CommentTypeComment:
		title += " 有新评论"
		content = commentInfo.Content + "<hr/>" + content
	case service5.CommentTypeOpen:
		title += " 打开"
	case service5.CommentTypeClose:
		title += " 关闭"
	default:
		return nil
	}
	for _, uid := range list {
		u, err := n.userSvc.SelfInfo(uid)
		if err != nil {
			continue
		}
		uc, err := n.userSvc.GetUserConfig(u.UID)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyScriptIssueComment].(bool); ok && !n {
			continue
		}
		_ = n.notifySvc.NotifyEmailFrom(user.Username, u.Email, title, content, "text/html")
	}
	// 解析是否有艾特并通知
	users, err := n.parseContent(commentInfo.Content)
	if err != nil {
		logrus.Errorf("parseContent: %v", err)
		return nil
	}
	for _, v := range users {
		uc, err := n.userSvc.GetUserConfig(v.UID)
		if err != nil {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyAt].(bool); ok && !n {
			continue
		}
		if n, ok := uc.Notify[service3.UserNotifyScriptIssueComment].(bool); ok && !n {
			continue
		}
		_ = n.notifySvc.NotifyEmailFrom(user.Username, v.Email, user.Username+" 在 "+issueInfo.Title+" 中有提及到您", content, "text/html")
	}
	return nil
}

// 解析内容查看是否有艾特的人,返回用户信息
func (n *ScriptSubscriber) parseContent(content string) ([]*respond.User, error) {
	r := regexp.MustCompile("@(\\S+)")
	list := r.FindAllStringSubmatch(content, -1)
	var users []*respond.User
	for _, v := range list {
		user, err := n.userSvc.FindByUsername(v[1], true)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}
