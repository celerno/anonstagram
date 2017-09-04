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

	"github.com/revel/examples/upload/app/routes"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2"
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
	tags     string
	path     string
	shareURL string
}
type dbComment struct {
	id       bson.ObjectId `bson:"_id,omintempy"`
	comment  string
	shareURL string
}

var awsKey string
var awsSecret string

func (a App) Index() revel.Result {
	saludo := "eea ea eaeaeaea"

	url := "mongodb://heroku_c8zbgw18:rv1q4bsr8036m2igtqhs9q22ro@ds117919.mlab.com:17919/heroku_c8zbgw18"
	session, err := mgo.Dial(url)
	c := session.DB("heroku_c8zbgw18").C("anonstagram")
	count, _ := c.Count()
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
		return c.Redirect(routes.App.Before())
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
	return c.RenderJSON(FileInfo{
		ContentType: c.Params.Files["pic"][0].Header.Get("Content-Type"),
		Filename:    awsutil.StringValue(resp),
		RealFormat:  format,
		Resolution:  fmt.Sprintf("%dx%d", conf.Width, conf.Height),
		Size:        len(pic),
		Status:      "Successfully uploaded",
	})
}
