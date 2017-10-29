package controllers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	
	"github.com/revel/revel"

)

const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
)

type FileInfo struct {
	ContentType string
	Filename    string
	RealFormat  string `json:",omitempty"`
	Resolution  string `json:",omitempty"`
	Size        int
	Status      string `json:",omitempty"`
}

type App struct {
	*revel.Controller
}

type dbPost struct {
	id       bson.ObjectId `bson:"_id,omintempy"`
	Tags     string `json:"tags"`
	Path     string `json:"path"`
	ShareURL string `json:"shareURL"`
}
type dbComment struct {
	id       bson.ObjectId `bson:"_id,omintempy"`
	comment  string
	shareURL string
}

var awsKey string
var awsSecret string
var dbC mgo.Collection

func (a App) Index() revel.Result {
	saludo := "eea ea eaeaeaea"
	mongoUser :=	os.Getenv("mongoUser")
	mongoURL := os.Getenv("mongoURL")
	mongoSession, err := mgo.Dial(mongoURL)
	dbC := mongoSession.DB(mongoUser).C("anonstagram")
	count, _ := dbC.Count()
	saludo = saludo + ". " + fmt.Sprintf("%d", count)

	if err != nil {
		panic(err.Error)
	}

	return a.Render(saludo)
}

func (c *App) Upload(pic []byte) revel.Result {
	// Validation rules.
	log.Printf("bytesize %d", len(pic))
	c.Validation.Required(pic)
	c.Validation.MinSize(pic, 100*KB).
		Message("Minimum a file size of 100KB expected")
	c.Validation.MaxSize(pic, 5*MB).
		Message("File cannot be larger than 2MB")
	
	// Check format of the file.
	conf, format, err := image.DecodeConfig(bytes.NewReader(pic))
	c.Validation.Required(err == nil).Key("pic").Message("Incorrect file format")
	c.Validation.Required(format == "jpeg" || format == "png").Key("pic").
		Message("JPEG or PNG file format is expected")

	// Check resolution.
	c.Validation.Required(conf.Height >= 150 && conf.Width >= 150).Key("pic").
		Message("Minimum allowed resolution is 150x150px")

	// Handle errors.
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	awsKey = os.Getenv("awsKey")
	awsSecret = os.Getenv("awsSecret")
	token := ""
	log.Println(awsKey)
	log.Println(awsSecret)
	creds := credentials.NewStaticCredentials(awsKey, awsSecret, token)
	_, errAWS := creds.Get()
	if errAWS != nil {
		panic(errAWS.Error())
	}
	cfg := aws.NewConfig().WithRegion("us-west-1").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
	log.Println(svc.ClientInfo)
	params := &s3.PutObjectInput{
		Bucket:        aws.String("anonstagram"),
		Key:           aws.String("/" + c.Params.Files["pic"][0].Filename),
		Body:          bytes.NewReader(pic),
		ContentLength: aws.Int64(int64(len(pic))),
		ContentType:   aws.String("image/" + format),
	}

	resp, errPut := svc.PutObject(params)
	if errPut != nil {
		panic(errPut.Error())
	}

	mongoUser :=	os.Getenv("mongoUser")
	mongoURL := os.Getenv("mongoURL")
	mongoSession, err := mgo.Dial(mongoURL)
	if err==nil {
		dbC := mongoSession.DB(mongoUser).C("anonstagram")
		var post dbPost
		post.Path = awsutil.StringValue(resp)
		post.ShareURL = "/anon/" + awsutil.StringValue(resp)
		post.Tags = "lame, very lame"

		dbC.Insert(post)
		
	}

	return c.RenderJSON(FileInfo{
		ContentType: c.Params.Files["pic"][0].Header.Get("Content-Type"),
		Filename:    awsutil.StringValue(resp),
		RealFormat:  format,
		Resolution:  fmt.Sprintf("%dx%d", conf.Width, conf.Height),
		Size:        len(pic),
		Status:      "Successfully uploaded",
	})
}