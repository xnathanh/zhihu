package models

import (
	"fmt"
	"html/template"
	"log"
	"time"
)

type User struct {
	ID        uint   `json:"id"`
	Email     string `json:"-"` //username
	Name      string `json:"fullname"`
	Password  string `json:"-"`
	URLToken  string `json:"url_token"`
	Gender    int    `json:"gender"`
	Headline  string `json:"headline"`
	AvatarURL string `json:"avatar_url"`

	Posts         []*Post
	AnswerCount   uint `json:"answer_count"`
	FollowerCount uint `json:"follower_count"`

	Followed  bool `json:"is_followed"` //followed by user who sent request
	Following bool `json:"is_following"`
	Anonymous bool `json:"is_anonymous"`
}

type Member struct {
	User
	MarkedCount    uint `json:"marked_count"`
	FollowingCount uint `json:'following_count"`
	//	Educations []*Education
	PrivacyProtected bool `json:"is_privacy_protected"`
	VoteupCount      uint `json:"voteup_count"`
	//Blocked bool `json:"is_blocked"`
	Description         string `json:"description"`
	FollowingTopicCount uint   `json:"following_topic_count"`
	ThankedCount        uint   `json:"thanked_count"`
}

const (
	AskQuestionAction = iota
	AnswerQuestionAction
	FollowQuestionAction
	VoteupAnswerAction
	OtherAction
)

type Action struct {
	*User       `json:"user"`
	Type        int `json:"type"`
	*Question   `json:"question"`
	*Answer     `json:"answer"`
	DateCreated string `json:"created_at"`
}

type Question struct {
	ID           string `json:"id"`
	User         *User  `json:"user"`
	Title        string `json:"title"`
	Detail       string `json:"detail"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`

	VisitCount    uint `json:"visit_count"`
	AnswerCount   uint `json:"answer_count"`
	CommentCount  uint `json:"comment_count"`
	FollowerCount uint `json:"follower_count"`

	Topics         []*Topic
	TopicURLTokens []string `json:"topic_url_tokens"`
	Answers        []*Answer
	//	Comments  []*QuestionComment
	//	Followers []*User

	Followed             bool `json:"is_followed"` //followed by user who sent request
	Answered             bool `json:"is_answered"`
	VisitorAnswerID      uint `json:"visitor_answer_id"`
	VisitorAnswerDeleted bool `json:"visitor_answer_deleted"`
	Anonymous            bool `json:"is_anonymous"`
}

type Topic struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Answer struct {
	ID string `json:"id"`
	*Question
	Author       *User `json:"author"`
	Content      template.HTML
	DateCreated  string
	DateModified string

	MarkedCount   uint
	UpvoteCount   uint
	DownvoteCount uint
	CommentCount  uint

	Comments []*AnswerComment
	//	Upvotes  []*User

	Upvoted   bool
	Downvoted bool
	IsAuthor  bool
	Deleted   bool
}

type Post struct {
}

type Comment struct {
	ID            uint   `json:"id"`
	Author        *User  `json:"author"`
	DateCreated   string `json:"date_created"`
	UpvoteCount   uint   `json:"upvote_count"`
	DownvoteCount uint   `json:"downvote_count"`
	Content       string `json:"content"`
	LikeCount     uint   `json:"like_count"`
	Liked         bool   `json:"is_liked"`
}

type AnswerComment struct {
	*Answer
	Comment
}

type QuestionComment struct {
	*Question
	Comment
}

func NewQuestion() *Question {
	question := new(Question)
	question.User = new(User)
	return question
}

func NewAnswer() *Answer {
	answer := new(Answer)
	answer.Question = NewQuestion()
	answer.Author = new(User)
	return answer
}

func HandleNewAction(user uint, action int, id string) {
	var err error
	defer func() {
		if err != nil {
			log.Println("models.HandleNewAction(): ", err)
		}
	}()
	conn := redisPool.Get()
	rows, err := db.Query("SELECT follower_id FROM member_followers WHERE member_id=?", user)
	if err != nil {
		return
	}
	now := time.Now().Unix()
	key := fmt.Sprintf("profile:%d", user)
	field := fmt.Sprintf("%d:%s", action, id)
	conn.Do("ZADD", key, now, field)
	var fid string
	for rows.Next() {
		if err = rows.Scan(&fid); err != nil {
			continue
		}
		key := "home:" + fid
		field := fmt.Sprintf("%d:%d:%s", user, action, id)
		conn.Do("ZADD", key, now, field)
	}
	if action == VoteupAnswerAction {
		n, err := conn.Do("SCARD", "upvoted:"+id)
		if err != nil {
			return
		}
		score := 86400 + n.(int64)*432
		println(id, score)
		_, err = conn.Do("ZADD", "rank", score, id)
		if err != nil {
			return
		}
	}
}

func RemoveAction(user uint, action int, id string) {
	var err error
	defer func() {
		if err != nil {
			log.Println("models.DeleteAction(): ", err)
		}
	}()

	conn := redisPool.Get()
	rows, err := db.Query("SELECT follower_id FROM member_followers WHERE member_id=?", user)
	if err != nil {
		return
	}
	key := fmt.Sprintf("profile:%d", user)
	member := fmt.Sprintf("%d:%s", action, id)
	conn.Do("ZREM", key, member)
	var fid string
	for rows.Next() {
		if err = rows.Scan(&fid); err != nil {
			continue
		}
		key := "home:" + fid
		member := fmt.Sprintf("%d:%d:%s", user, action, id)
		conn.Do("ZREM", key, member)
	}
}
