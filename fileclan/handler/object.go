package handler

import (
	"github.com/qinguanri/fileclan/middlewares"
	"fmt"
	"github.com/valyala/fasthttp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

/*
 * metadata是文件的描述信息。
 * 源码文件和二进制文件存在一对一的关系
 */
type MetaData struct {
	FileName  string    // 源码文件的名称
	CommitId  string    // 源码文件在git上的commitid
	Path      string    // 源码文件在git上的文件路径
	FileMd5   string    // 源码文件的md5
	CreatedAt time.Time // 创建时间
}

/*
 * FileObject是一个要存储到mongodb的文件对象
 */
type FileObject struct {
	Meta MetaData // 文件的描述信息
	Data []byte   // 二进制文件的内容
}

// Hello接口用于探活
func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "FileClan API Server is OK!")
}

func PutObject(ctx *fasthttp.RequestCtx) {
	filename := string(ctx.URI().QueryArgs().Peek("filename"))
	commitid := string(ctx.URI().QueryArgs().Peek("commitid"))
	path := string(ctx.URI().QueryArgs().Peek("path"))
	filemd5 := string(ctx.URI().QueryArgs().Peek("md5"))

	if filename == "" || commitid == "" || path == "" {
		ctx.Error("invalid query arguments.", 400)
		return
	}

	body := ctx.PostBody()

	var file = &FileObject{
		Meta: MetaData{
			FileName:  filename,
			CommitId:  commitid,
			Path:      path,
			FileMd5:   filemd5,
			CreatedAt: time.Now(),
		},
		Data: body,
	}

	db := middlewares.Backend.Db.Copy()
	defer db.Close()
	c := db.DB("fileclan").C("binaryfile")

	index := mgo.Index{
		Key: []string{
			"Meta.FileName",
			"Meta.FileMd5",
		},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		ctx.Error("ensure index failed when put object.", 503)
		return
	}

	err = c.Insert(file)

	if err != nil {
		ctx.Error("insert into mongo failed when put object.", 503)
		return
	}
}

func HeadObject(ctx *fasthttp.RequestCtx) {
	md5 := string(ctx.URI().QueryArgs().Peek("md5"))
	filename := string(ctx.URI().QueryArgs().Peek("filename"))
	if md5 == "" || filename == "" {
		ctx.Error("invalid query arguments.md5 or filename is nil", 400)
		return
	}

	db := middlewares.Backend.Db.Copy()
	defer db.Close()
	c := db.DB("fileclan").C("binaryfile")

	fileObject := FileObject{}
	err := c.Find(bson.M{"meta.sourcefilemd5": md5, "meta.filename": filename}).One(&fileObject)

	if err != nil {
		ctx.Error("object not found.", 404)
		return
	}

	ctx.Response.Header.Set("FileName", fileObject.Meta.FileName)
	ctx.Response.Header.Set("CommitId", fileObject.Meta.CommitId)
	ctx.Response.Header.Set("Path", fileObject.Meta.Path)
	ctx.Response.Header.Set("FileMd5", fileObject.Meta.FileMd5)
	ctx.Response.Header.Set("CreatedAt", fileObject.Meta.CreatedAt.Format("2006-01-02 15:04:05"))
	ctx.Response.SetConnectionClose()
}

func GetObject(ctx *fasthttp.RequestCtx) {
	md5 := string(ctx.URI().QueryArgs().Peek("md5"))
	filename := string(ctx.URI().QueryArgs().Peek("filename"))
	if md5 == "" || filename == "" {
		ctx.Error("invalid query arguments.md5 or filename is nil", 400)
		return
	}

	db := middlewares.Backend.Db.Copy()
	defer db.Close()
	c := db.DB("fileclan").C("binaryfile")

	var file FileObject
	err := c.Find(bson.M{"meta.filemd5": md5, "meta.filename": filename}).One(&file)
	if err != nil {
		ctx.Error("object not found", 404)
		return
	}

	ctx.SetContentType("application/octet-stream")
	ctx.Response.Header.Set("Content-Transfer-Encoding", "binary")
	ctx.Response.Header.Set("Content-Disposition", "attachment; filename="+file.Meta.FileName)
	ctx.SetBody(file.Data)
}

func DeleteObject(ctx *fasthttp.RequestCtx) {
	md5 := string(ctx.URI().QueryArgs().Peek("md5"))
	filename := string(ctx.URI().QueryArgs().Peek("filename"))
	if md5 == "" || filename == "" {
		ctx.Error("invalid query arguments.md5 or filename is nil", 400)
		return
	}

	db := middlewares.Backend.Db.Copy()
	defer db.Close()
	c := db.DB("fileclan").C("binaryfile")

	c.Remove(bson.M{"meta.filemd5": md5, "meta.filename": filename})
}
