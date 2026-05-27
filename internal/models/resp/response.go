package resp

import (
	"bbs-go/internal/models/constants"
	"time"

	"github.com/mlogclub/simple/web"
)

// UserInfo 用户简单信息
type UserInfo struct {
	Id           string           `json:"id"`
	Nickname     string           `json:"nickname"`
	Avatar       string           `json:"avatar"`
	SmallAvatar  string           `json:"smallAvatar"`
	Gender       constants.Gender `json:"gender"`
	Birthday     *time.Time       `json:"birthday"`
	TopicCount   int              `json:"topicCount"`   // 话题数量
	CommentCount int              `json:"commentCount"` // 跟帖数量
	FansCount    int              `json:"fansCount"`    // 粉丝数量
	FollowCount  int              `json:"followCount"`  // 关注数量
	Score        int              `json:"score"`        // 积分
	Exp          int              `json:"exp"`          // 经验值
	Level        int              `json:"level"`        // 等级
	LevelTitle   string           `json:"levelTitle"`   // 等级称号
	Description  string           `json:"description"`
	CreateTime   int64            `json:"createTime"`

	Forbidden bool `json:"forbidden"` // 是否禁言
	Followed  bool `json:"followed"`  // 是否关注

	// ExpProgress 经验值进度（当前等级内进度条数据），由 BuildUserInfo 根据 LevelConfig 计算填充；未登录或异常时为 nil
	ExpProgress *ExpProgressResponse `json:"expProgress,omitempty"`
}

// ExpProgressResponse 用户经验值进度（用于当前等级内的进度条展示）
type ExpProgressResponse struct {
	CurrentExp          int    `json:"currentExp"`
	Level               int    `json:"level"`
	LevelTitle          string `json:"levelTitle"`
	ExpInCurrentLevel   int    `json:"expInCurrentLevel"`
	ExpNeedForNextLevel int    `json:"expNeedForNextLevel"`
	ExpProgressPercent  int    `json:"expProgressPercent"`
	IsMaxLevel          bool   `json:"isMaxLevel"`
}

// UserDetail 用户详细信息
type UserDetail struct {
	UserInfo
	Username             string `json:"username"`
	BackgroundImage      string `json:"backgroundImage"`
	SmallBackgroundImage string `json:"smallBackgroundImage"`
	HomePage             string `json:"homePage"`
	Status               int    `json:"status"`
}

// UserProfile 用户个人信息
type UserProfile struct {
	UserDetail
	Roles         []string `json:"roles"`
	Permissions   []string `json:"permissions"`
	PasswordSet   bool     `json:"passwordSet"` // 密码已设置
	Email         string   `json:"email"`
	EmailVerified bool     `json:"emailVerified"`
}

type TagResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ArticleSimpleResponse struct {
	Id           int64          `json:"id"`
	User         *UserInfo      `json:"user"`
	Tags         *[]TagResponse `json:"tags"`
	Title        string         `json:"title"`
	Summary      string         `json:"summary"`
	Cover        *ImageInfo     `json:"cover"`
	SourceUrl    string         `json:"sourceUrl"`
	ViewCount    int64          `json:"viewCount"`
	CommentCount int64          `json:"commentCount"`
	LikeCount    int64          `json:"likeCount"`
	CreateTime   int64          `json:"createTime"`
	Status       int            `json:"status"`
	Favorited    bool           `json:"favorited"`
}

type ArticleResponse struct {
	ArticleSimpleResponse
	Content string         `json:"content"`
	Toc     []TopicTocItem `json:"toc,omitempty"`
}

type CategoryResponse struct {
	Id          int64                  `json:"id"`
	ParentId    int64                  `json:"parentId"` // 父节点ID，0=一级
	Name        string                 `json:"name"`
	Type        constants.CategoryType `json:"type"`
	Logo        string                 `json:"logo"`
	Description string                 `json:"description"`
	Children    []CategoryResponse     `json:"children,omitempty"` // 子节点（发帖可选时用）
}

// CategoryTreeItem 后台节点树形列表项
type CategoryTreeItem struct {
	Id          int64                  `json:"id"`
	ParentId    int64                  `json:"parentId"`
	Name        string                 `json:"name"`
	Type        constants.CategoryType `json:"type"`
	Logo        string                 `json:"logo"`
	Description string                 `json:"description"`
	SortNo      int                    `json:"sortNo"`
	Status      int                    `json:"status"`
	CreateTime  int64                  `json:"createTime"`
	Children    []CategoryTreeItem     `json:"children"`
}

type SearchTopicResponse struct {
	Id         int64             `json:"id"`
	User       *UserInfo         `json:"user"`
	Category   *CategoryResponse `json:"category"`
	Tags       *[]TagResponse    `json:"tags"`
	Title      string            `json:"title"`
	Summary    string            `json:"summary"`
	CreateTime int64             `json:"createTime"`
}

type SearchArticleResponse struct {
	Id         int64          `json:"id"`
	User       *UserInfo      `json:"user"`
	Tags       *[]TagResponse `json:"tags"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	CreateTime int64          `json:"createTime"`
}

type SearchUserResponse struct {
	User        *UserInfo `json:"user"`
	Nickname    string    `json:"nickname"`
	Username    string    `json:"username"`
	Description string    `json:"description"`
	CreateTime  int64     `json:"createTime"`
}

type TopicTocItem struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Level int    `json:"level"`
}

// 帖子列表返回实体
type TopicResponse struct {
	Id                string               `json:"id"`
	Type              constants.TopicType  `json:"type"`
	QaStatus          constants.QaStatus   `json:"qaStatus"`
	AcceptedCommentId int64                `json:"acceptedCommentId"`
	SolvedAt          int64                `json:"solvedAt"`
	BountyScore       int                  `json:"bountyScore"`
	User              *UserInfo            `json:"user"`
	Category          *CategoryResponse    `json:"category"`
	Tags              *[]TagResponse       `json:"tags"`
	Title             string               `json:"title"`
	Summary           string               `json:"summary"`
	Content           string               `json:"content"`
	Toc               []TopicTocItem       `json:"toc,omitempty"`
	ImageList         []ImageInfo          `json:"imageList"`
	LastCommentTime   int64                `json:"lastCommentTime"`
	ViewCount         int64                `json:"viewCount"`
	CommentCount      int64                `json:"commentCount"`
	LikeCount         int64                `json:"likeCount"`
	Liked             bool                 `json:"liked"`
	CreateTime        int64                `json:"createTime"`
	Recommend         bool                 `json:"recommend"`
	RecommendTime     int64                `json:"recommendTime"`
	Sticky            bool                 `json:"sticky"`
	StickyTime        int64                `json:"stickyTime"`
	Status            int                  `json:"status"`
	Favorited         bool                 `json:"favorited"`
	IpLocation        string               `json:"ipLocation"`
	Vote              *VoteResponse        `json:"vote"`
	Attachments       []AttachmentResponse `json:"attachments,omitempty"`
}

type AttachmentResponse struct {
	Id            string `json:"id"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize"`
	DownloadScore int    `json:"downloadScore"`
	DownloadCount int    `json:"downloadCount"`
	Downloaded    bool   `json:"downloaded"`
}

type VoteResponse struct {
	Id                   int64                          `json:"id"`
	Type                 constants.VoteType             `json:"type"`
	PollType             constants.PollType             `json:"pollType"`
	Title                string                         `json:"title"`
	ExpiredAt            int64                          `json:"expiredAt"`
	VoteNum              int                            `json:"voteNum"`
	OptionCount          int                            `json:"optionCount"`
	VoteCount            int                            `json:"voteCount"`
	Expired              bool                           `json:"expired"`
	Voted                bool                           `json:"voted"`
	CanViewResults       bool                           `json:"canViewResults"`
	Anonymous            bool                           `json:"anonymous"`
	HideResults          constants.HideResultsType      `json:"hideResults"`
	ClosedAt             *int64                         `json:"closedAt"`
	StanceReasonRequired constants.StanceReasonRequired `json:"stanceReasonRequired"`
	OptionIds            []int64                        `json:"optionIds"`
	Options              []VoteOptionResponse           `json:"options"`
	Outcome              *OutcomeResponse               `json:"outcome,omitempty"`
}

type VoteOptionResponse struct {
	Id         int64   `json:"id"`
	Content    string  `json:"content"`
	SortNo     int     `json:"sortNo"`
	VoteCount  int     `json:"voteCount"`
	Percent    float64 `json:"percent"`
	Voted      bool    `json:"voted"`
	Meaning    string  `json:"meaning,omitempty"`
	Prompt     string  `json:"prompt,omitempty"`
	TotalScore int     `json:"totalScore,omitempty"`
	VoterCount int     `json:"voterCount,omitempty"`
}

type StanceResponse struct {
	Id             int64                  `json:"id"`
	PollId         int64                  `json:"pollId"`
	ParticipantId  int64                  `json:"participantId"`
	Participant    *UserInfo              `json:"participant,omitempty"`
	Reason         string                 `json:"reason"`
	ReasonFormat   string                 `json:"reasonFormat"`
	Latest         bool                   `json:"latest"`
	CastAt         *int64                 `json:"castAt"`
	RevokedAt      *int64                 `json:"revokedAt"`
	NoneOfTheAbove bool                   `json:"noneOfTheAbove"`
	Choices        []StanceChoiceResponse `json:"choices,omitempty"`
	CreateTime     int64                  `json:"createTime"`
}

type StanceChoiceResponse struct {
	Id           int64 `json:"id"`
	StanceId     int64 `json:"stanceId"`
	PollOptionId int64 `json:"pollOptionId"`
	Score        int   `json:"score"`
}

type OutcomeResponse struct {
	Id              int64               `json:"id"`
	PollId          int64               `json:"pollId"`
	Statement       string              `json:"statement"`
	StatementFormat string              `json:"statementFormat"`
	AuthorId        int64               `json:"authorId"`
	Author          *UserInfo           `json:"author,omitempty"`
	PollOptionId    *int64              `json:"pollOptionId"`
	PollOption      *VoteOptionResponse `json:"pollOption,omitempty"`
	Latest          bool                `json:"latest"`
	ReviewOn        *int64              `json:"reviewOn"`
	CreateTime      int64               `json:"createTime"`
	UpdateTime      int64               `json:"updateTime"`
}

type CommentResponse struct {
	Id           int64                 `json:"id"`
	User         *UserInfo             `json:"user"`
	EntityType   string                `json:"entityType"`
	EntityId     int64                 `json:"entityId"`
	ContentType  constants.ContentType `json:"contentType"`
	Content      string                `json:"content"`
	ImageList    []ImageInfo           `json:"imageList"`
	LikeCount    int64                 `json:"likeCount"`
	CommentCount int64                 `json:"commentCount"`
	Liked        bool                  `json:"liked"`
	QuoteId      int64                 `json:"quoteId"`
	Quote        *CommentResponse      `json:"quote"`
	Replies      *web.CursorResult     `json:"replies"`
	IpLocation   string                `json:"ipLocation"`
	Status       int                   `json:"status"`
	CreateTime   int64                 `json:"createTime"`
}

type FavoriteResponse struct {
	Id         int64     `json:"id"`
	EntityType string    `json:"entityType"`
	EntityId   int64     `json:"entityId"`
	Deleted    bool      `json:"deleted"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	User       *UserInfo `json:"user"`
	Url        string    `json:"url"`
	CreateTime int64     `json:"createTime"`
}

type MessageResponse struct {
	Id           int64     `json:"id"`
	From         *UserInfo `json:"from"`
	UserId       int64     `json:"userId"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	QuoteContent string    `json:"quoteContent"`
	Type         int       `json:"type"`
	DetailUrl    string    `json:"detailUrl"`
	ExtraData    string    `json:"extraData"`
	Status       int       `json:"status"`
	CreateTime   int64     `json:"createTime"`
}

type ImageInfo struct {
	Url     string `json:"url"`
	Preview string `json:"preview"`
}

type TreeNode struct {
	Id       int64      `json:"id"`
	Key      int64      `json:"key"`
	Title    string     `json:"title"`
	Children []TreeNode `json:"children"`
}

type MenuResponse struct {
	Id         int64  `json:"id"`
	ParentId   *int64 `json:"parentId"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	SortNo     int    `json:"sortNo"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
}

type MenuTreeResponse struct {
	MenuResponse
	Level    int                `json:"level"`
	Children []MenuTreeResponse `json:"children"`
}

type DictResponse struct {
	Id         int64  `json:"id"`
	TypeId     int64  `json:"typeId"`
	ParentId   *int64 `json:"parentId"`
	Name       string `json:"name"`
	Label      string `json:"label"`
	Value      string `json:"value"`
	SortNo     int    `json:"sortNo"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
}

type DictListResponse struct {
	DictResponse
	Children []DictListResponse `json:"children"`
}

type TaskGroupInfo struct {
	Key  constants.TaskGroup `json:"key"`
	Name string              `json:"name"`
}

type TaskResponse struct {
	Id             int64                 `json:"id"`
	GroupName      constants.TaskGroup   `json:"groupName"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	EventType      string                `json:"eventType"`
	Period         constants.TaskPeriod  `json:"period"`
	EventCount     int                   `json:"eventCount"`
	MaxFinishCount int                   `json:"maxFinishCount"`
	Score          int                   `json:"score"`
	Exp            int                   `json:"exp"`
	BadgeId        int64                 `json:"badgeId"`
	BtnName        string                `json:"btnName"`
	ActionUrl      string                `json:"actionUrl"`
	SortNo         int                   `json:"sortNo"`
	StartTime      int64                 `json:"startTime"`
	EndTime        int64                 `json:"endTime"`
	Status         int                   `json:"status"`
	UserProgress   *TaskProgressResponse `json:"userProgress,omitempty"`
}

type TaskProgressResponse struct {
	PeriodKey      int `json:"periodKey"`
	EventProgress  int `json:"eventProgress"`
	EventTarget    int `json:"eventTarget"`
	FinishedCount  int `json:"finishedCount"`
	MaxFinishCount int `json:"maxFinishCount"`
}

type BadgeResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortNo      int    `json:"sortNo"`
	Status      int    `json:"status"`
	Owned       bool   `json:"owned"`
	Worn        bool   `json:"worn"`
	ObtainTime  int64  `json:"obtainTime"`
}

type CheckInResponse struct {
	Id              int64     `json:"id"`
	UserId          int64     `json:"userId"`
	LatestDayName   int       `json:"latestDayName"`
	ConsecutiveDays int       `json:"consecutiveDays"`
	CheckIn         bool      `json:"checkIn"`
	UpdateTime      int64     `json:"updateTime"`
	User            *UserInfo `json:"user,omitempty"`
}

type SiteNavResponse struct {
	Title           string            `json:"title"`
	Url             string            `json:"url"`
	OpenInNewWindow bool              `json:"openInNewWindow"`
	Children        []SiteNavResponse `json:"children,omitempty"`
}
