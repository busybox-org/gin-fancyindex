package main

import (
	"context"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/gin-fancyindex/internal"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var conf = new(internal.Config)

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
	ok, err := internal.PathExists(conf.Root)
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
		Handler:      internal.Router(conf), // Pass our instance of gin in.
	}
	go func() {
		logrus.Infof("Server is running on %s", srv.Addr)
		if err = srv.ListenAndServe(); err != nil {
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
