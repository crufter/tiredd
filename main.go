package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	m3o "github.com/micro/services/clients/go"
	db "github.com/micro/services/clients/go/db"
	user "github.com/micro/services/clients/go/user"
	uuid "github.com/satori/go.uuid"
)

var client = m3o.NewClient(os.Getenv("MICRO_API_TOKEN"))

// Types

type Post struct {
	Id           string  `json:"id"`
	UserId       string  `json:"userId"`
	UserName     string  `json:"userName"`
	Content      string  `json:"content"`
	Created      string  `json:"created"`
	Upvotes      float32 `json:"upvotes"`
	Downvotes    float32 `json:"downvotes"`
	Score        float32 `json:"score"`
	Title        string  `json:"title"`
	Url          string  `json:"url"`
	Sub          string  `json:"sub"`
	CommentCount float32 `json:"commentCount"`
}

type Comment struct {
	Content   string  `json:"content"`
	Parent    string  `json:"sub"`
	Upvotes   float32 `json:"upvotes"`
	Downvotes float32 `json:"downvotes"`
	PostId    string  `json:"postId"`
	UserName  string  `json:"usernName"`
	UserId    string  `json:"userId"`
}

type PostRequest struct {
	Post      Post   `json:"post"`
	SessionID string `json:"sessionId"`
}

type VoteRequest struct {
	Id        string `json:"id"`
	SessionID string `json:"sessionId"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CommentRequest struct {
	Comment   Comment `json:"comment"`
	SessionID string  `json:"sessionId"`
}

type CommentsRequest struct {
	PostId string `json:"postId"`
}

type PostsRequest struct {
	Min int32  `json:"min"`
	Sub string `json:"sub"`
}

// Endpoints

// upvote or downvote a post or a comment
func vote(w http.ResponseWriter, req *http.Request, upvote bool, isComment bool, t VoteRequest) error {
	if t.Id == "" {
		return fmt.Errorf("missing post id")
	}
	table := "posts"
	if isComment {
		table = "comments"
	}
	rsp, err := client.DbService.Read(&db.ReadRequest{
		Table: table,
		Id:    t.Id,
	})
	if err != nil {
		return err
	}
	if len(rsp.Records) == 0 {
		return fmt.Errorf("post or comment not found")
	}

	// auth
	sessionRsp, err := client.UserService.ReadSession(&user.ReadSessionRequest{
		SessionId: t.SessionID,
	})
	if err != nil {
		return err
	}
	if sessionRsp.Session.UserId == "" {
		return fmt.Errorf("user id not found")
	}

	// prevent double votes
	checkTable := table + "votecheck"
	checkId := t.Id + sessionRsp.Session.UserId
	checkRsp, err := client.DbService.Read(&db.ReadRequest{
		Table: checkTable,
		Id:    checkId,
	})
	if err == nil && (checkRsp != nil && len(checkRsp.Records) > 0) {
		return fmt.Errorf("already voted")
	}

	_, err = client.DbService.Create(&db.CreateRequest{
		Table: checkTable,
		Record: map[string]interface{}{
			"id": checkId,
		},
	})
	if err != nil {
		return err
	}

	obj := rsp.Records[0]
	key := "upvotes"
	if !upvote {
		key = "downvotes"
	}

	if _, ok := obj["upvotes"].(float64); !ok {
		obj["upvotes"] = float64(0)
	}
	if _, ok := obj["downvotes"].(float64); !ok {
		obj["downvotes"] = float64(0)
	}

	obj[key] = obj[key].(float64) + 1
	obj["score"] = obj["upvotes"].(float64) - obj["downvotes"].(float64)

	_, err = client.DbService.Update(&db.UpdateRequest{
		Table:  table,
		Id:     t.Id,
		Record: obj,
	})
	return err
}

func voteWrapper(upvote bool, isComment bool) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if cors(w, req) {
			return
		}

		decoder := json.NewDecoder(req.Body)
		var t VoteRequest
		err := decoder.Decode(&t)
		if err != nil {
			respond(w, nil, err)
			return
		}
		err = vote(w, req, upvote, isComment, t)
		respond(w, nil, err)
	}
}

func login(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t LoginRequest
	err := decoder.Decode(&t)
	if err != nil {
		respond(w, err, err)
		return
	}
	_, err = client.UserService.Read(&user.ReadRequest{
		Username: t.Username,
	})
	if err != nil {
		createRsp, err := client.UserService.Create(&user.CreateRequest{
			Username: t.Username,
			Email:    t.Username + "@" + t.Username + ".com",
			Password: t.Password,
		})
		if err != nil {
			respond(w, createRsp, err)
			return
		}
	}
	logRsp, err := client.UserService.Login(&user.LoginRequest{
		Username: t.Username,
		Password: t.Password,
	})
	respond(w, logRsp, err)
}

func readSession(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t user.ReadSessionRequest
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err.Error()))
	}
	rsp, err := client.UserService.ReadSession(&t)
	if err != nil {
		respond(w, rsp, err)
		return
	}
	readRsp, err := client.UserService.Read(&user.ReadRequest{
		Id: rsp.Session.UserId,
	})
	respond(w, map[string]interface{}{
		"session": rsp.Session,
		"account": readRsp.Account,
	}, err)
}

func post(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t PostRequest
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err.Error()))
	}
	userID := ""
	userName := ""
	if t.SessionID != "" {
		rsp, err := client.UserService.ReadSession(&user.ReadSessionRequest{
			SessionId: t.SessionID,
		})
		if err != nil {
			respond(w, rsp, err)
			return
		}
		userID = rsp.Session.UserId
		readRsp, err := client.UserService.Read(&user.ReadRequest{
			Id: userID,
		})
		if err != nil {
			respond(w, rsp, err)
			return
		}
		userName = readRsp.Account.Username
	}
	client.DbService.Create(&db.CreateRequest{
		Table: "posts",
		Record: map[string]interface{}{
			"id":        uuid.NewV4(),
			"userId":    userID,
			"userName":  userName,
			"content":   t.Post.Content,
			"url":       t.Post.Url,
			"upvotes":   float64(0),
			"downvotes": float64(0),
			"score":     float64(0),
			"sub":       t.Post.Sub,
			"title":     t.Post.Title,
			"created":   time.Now(),
		},
	})
}

func comment(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t CommentRequest
	err := decoder.Decode(&t)
	if err != nil {
		respond(w, nil, err)
		return
	}
	userID := ""
	userName := ""
	// get user if available
	if t.SessionID != "" {
		rsp, err := client.UserService.ReadSession(&user.ReadSessionRequest{
			SessionId: t.SessionID,
		})
		if err != nil {
			respond(w, rsp, err)
			return
		}
		userID = rsp.Session.UserId
		readRsp, err := client.UserService.Read(&user.ReadRequest{
			Id: userID,
		})
		if err != nil {
			respond(w, rsp, err)
			return
		}
		userName = readRsp.Account.Username
	}
	if t.Comment.PostId == "" {
		respond(w, nil, fmt.Errorf("no post id"))
		return
	}

	// get post to update comment counter
	readRsp, err := client.DbService.Read(&db.ReadRequest{
		Table: "posts",
		Id:    t.Comment.PostId,
	})
	if err != nil {
		respond(w, nil, err)
		return
	}
	if readRsp == nil || len(readRsp.Records) == 0 {
		respond(w, nil, fmt.Errorf("post not found"))
		return
	}
	if len(readRsp.Records) > 1 {
		respond(w, nil, fmt.Errorf("multiple posts found"))
		return
	}

	// create comment
	_, err = client.DbService.Create(&db.CreateRequest{
		Table: "comments",
		Record: map[string]interface{}{
			"id":        uuid.NewV4(),
			"userId":    userID,
			"userName":  userName,
			"content":   t.Comment.Content,
			"parent":    t.Comment.Parent,
			"postId":    t.Comment.PostId,
			"upvotes":   float64(0),
			"downvotes": float64(0),
			"score":     float64(0),
			"created":   time.Now(),
		},
	})
	if err != nil {
		respond(w, nil, err)
		return
	}

	// update counter
	oldCount, ok := readRsp.Records[0]["commentCount"].(float64)
	if !ok {
		oldCount = 0
	}
	oldCount++
	readRsp.Records[0]["commentCount"] = oldCount
	_, err = client.DbService.Update(&db.UpdateRequest{
		Table:  "posts",
		Id:     t.Comment.PostId,
		Record: readRsp.Records[0],
	})
	respond(w, nil, err)
}

func posts(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	var t PostsRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&t)
	r := &db.ReadRequest{
		Table:   "posts",
		Order:   "desc",
		OrderBy: "created",
	}
	query := ""
	if t.Min > 0 {
		query += "score > " + fmt.Sprintf("%v", t.Min)
	}
	if t.Sub != "all" && t.Sub != "" {
		if query != "" {
			query += " and "
		}
		query += fmt.Sprintf("sub == '%v'", t.Sub)
	}
	if query != "" {
		r.Query = query
	}

	rsp, err := client.DbService.Read(r)
	respond(w, rsp, err)
}

func comments(w http.ResponseWriter, req *http.Request) {
	if cors(w, req) {
		return
	}

	var t CommentsRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err.Error()))
	}
	rsp, err := client.DbService.Read(&db.ReadRequest{
		Table:   "comments",
		Order:   "desc",
		Query:   "postId == '" + t.PostId + "'",
		OrderBy: "created",
	})
	respond(w, rsp, err)
}

// Utils

func cors(w http.ResponseWriter, req *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func respond(w http.ResponseWriter, i interface{}, err error) {
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
	}
	if i == nil {
		i = map[string]interface{}{}
	}
	if err != nil {
		i = map[string]interface{}{
			"error": err.Error(),
		}
	}
	bs, _ := json.Marshal(i)
	fmt.Fprintf(w, fmt.Sprintf("%v", string(bs)))
}

func main() {
	http.HandleFunc("/upvotePost", voteWrapper(true, false))
	http.HandleFunc("/downvotePost", voteWrapper(false, false))
	http.HandleFunc("/upvoteComment", voteWrapper(true, true))
	http.HandleFunc("/downvoteComment", voteWrapper(false, true))
	http.HandleFunc("/posts", posts)
	http.HandleFunc("/post", post)
	http.HandleFunc("/comment", comment)
	http.HandleFunc("/comments", comments)
	http.HandleFunc("/login", login)
	http.HandleFunc("/readSession", readSession)

	http.ListenAndServe(":8090", nil)
}
