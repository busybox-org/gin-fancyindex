package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/gin-fancyindex"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Addr string `envconfig:"ADDR" default:""`
	Port string `envconfig:"PORT" default:"8080"`
	Root string `envconfig:"ROOT" default:"/share"`
	Auth bool   `envconfig:"AUTH" default:"false"`
	User string `envconfig:"USER" default:"admin"`
	Pass string `envconfig:"PASS" default:"admin"`
}

var conf = new(Config)

func init() {
	err := envconfig.Process("", conf)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339,
		DisableHTMLEscape: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			file = fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
			_f := strings.Split(frame.Function, ".")
			function = _f[len(_f)-1]
			return
		},
	})
}

func main() {
	ok, err := PathExists(conf.Root)
	if err != nil {
		logrus.Fatal(err)
	}
	if !ok {
		logrus.Fatal("root path not found")
	}

	logrus.Infoln("Starting server...")
	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%s", conf.Addr, conf.Port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      router(), // Pass our instance of gin in.
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.Error(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		logrus.Error(err)
		os.Exit(255)
	}
	os.Exit(0)
}

type accessLog struct {
	TimeStamp  string `json:"timestamp"`
	ClientIP   string `json:"client_ip"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Protocol   string `json:"protocol"`
	StatusCode int    `json:"status"`
	Latency    int64  `json:"duration"`
	BodySize   int    `json:"body_size"`
}

func router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
		cors.Default(),
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			log := &accessLog{
				TimeStamp:  param.TimeStamp.Format(time.RFC3339),
				ClientIP:   param.ClientIP,
				Method:     param.Method,
				Path:       param.Path,
				Protocol:   param.Request.Proto,
				StatusCode: param.StatusCode,
				Latency:    int64(param.Latency),
				BodySize:   param.BodySize,
			}
			bs, err := json.Marshal(log)
			if err != nil {
				logrus.Error(err)
				return ""
			}
			// your custom format
			return string(bs) + "\n"
		}))
	if conf.Auth {
		r.Use(gin.BasicAuth(gin.Accounts{
			conf.User: conf.Pass,
		}))
	}
	fancyindex.StaticFS(r, conf.Root)
	return r
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
