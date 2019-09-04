// Package fromenv provides utilities to create lile options from environment
// variables. fromenv will error with fatal if it cannot resolve or errors
package fromenv

import (
	"fmt"
	"os"

	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/google"
	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sirupsen/logrus"
)

func Tracer(name string) opentracing.Tracer {
	var rep reporter.Reporter
	var err error

	zipkinHost := os.Getenv("ZIPKIN_SERVICE_HOST")
	zipkinPort := os.Getenv("ZIPKIN_SERVICE_PORT")
	if zipkinHost != "" && zipkinPort != "" {
		addr := fmt.Sprintf("http://%s:%s/api/v1/spans", zipkinHost, zipkinPort)
		rep = zipkinHTTP.NewReporter(addr)
		logrus.Infof("Using Zipkin HTTP tracer: %s", addr)
	} else {
		logrus.Infof("Using Zipkin Global tracer")
		return opentracing.GlobalTracer()
	}

	// create tracer.
	nativeTracer, err := zipkin.NewTracer(
		rep,
		zipkin.WithTraceID128Bit(true),
	)
	if err != nil {
		logrus.Fatalf("unable to create Zipkin tracer: %+v", err)
	}
	tracer := zipkinot.Wrap(nativeTracer)

	// explicitly set our tracer to be the default tracer.
	opentracing.SetGlobalTracer(tracer)

	return tracer
}

func PubSubProvider() pubsub.Provider {
	gpid := os.Getenv("GOOGLE_PUBSUB_PROJECT_ID")
	if gpid != "" {
		gc, err := google.NewGoogleCloud(gpid)
		if err != nil {
			logrus.Fatalf("fronenv: Google Cloud pubsub err: %s", err)
			return nil
		}

		logrus.Infof("Using Google Cloud pubsub: %s", gpid)
		return gc
	}

	logrus.Warn("Using noop pubsub provider")
	return pubsub.NoopProvider{}
}
