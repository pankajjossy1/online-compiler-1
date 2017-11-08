package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type Build struct {
	Code     string `form:"code"`
	Language string `form:"language"`
}

type Run struct {
	Code     string `form:"code"`
	Language string `form:"language"`
	Stdin    string `form:"stdin"`
}

func main() {
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Docker client is not connected.")
	}
	options := types.ContainerListOptions{All: true}

	res, err := cli.ImagePull(ctx, "ugwis/online-compiler", types.ImagePullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, res)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/build", func(c *gin.Context) {
		var query Build
		if err := c.BindJSON(&query); err == nil {
			fmt.Printf("%v\n", query.Code)
			resp, err := cli.ContainerCreate(ctx, &container.Config{
				Image: "ugwis/online-compiler",
				Cmd:   strings.Split(query.Code, " "),
			}, nil, nil, "")
			if err != nil {
				log.Fatal(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			// Start container
			if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
				log.Fatal(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				log.Fatal(err)
			}
			c.Stream(func(w io.Writer) bool {
				io.Copy(w, out)
				return true
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})
	r.GET("/node", func(c *gin.Context) {
		containers, err := cli.ContainerList(ctx, options)
		if err != nil {
			log.Print(err)
			c.JSON(500, gin.H{
				"error": "Does not permit to fetch container list",
			})
		}
		c.JSON(200, gin.H{
			"containers": containers,
		})
	})
	r.Run()
}
