// Copyright 2021 The OCGI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"

	"k8s.io/klog"

	"github.com/ocgi/general-pod-autoscaler/pkg/util"
	"github.com/ocgi/general-pod-autoscaler/pkg/validator"
)

func Run(s *ServerRunOptions) error {
	stopCh := util.SetupSignalHandler()

	webHook := webhook.NewWebhookServer()

	// Start debug monitor.
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webHook.Serve)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", "ok")
	})

	server := &http.Server{
		Addr:         net.JoinHostPort(s.Address, strconv.Itoa(s.Port)),
		Handler:      mux,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	klog.V(1).Infof("listening on %v", server.Addr)
	if s.TlsCert != "" && s.TlsKey != "" {
		klog.V(1).Infof("using HTTPS service")
		tlsConfig, err := getTLSConfig(s)
		if err != nil {
			return err
		}
		server.TLSConfig = tlsConfig
		go func() {
			klog.Fatal(server.ListenAndServeTLS(s.TlsCert, s.TlsKey))
		}()
	} else {
		go func() {
			klog.V(1).Infof("using HTTP service")
			klog.Fatal(server.ListenAndServe())
		}()
	}

	select {
	case <-stopCh:
		klog.Info("http server received stop signal, waiting for all requests to finish")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}
	return nil
}

func getTLSConfig(s *ServerRunOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		//		Certificates: []tls.Certificate{cert},
		// Avoid fallback on insecure SSL protocols
		MinVersion: tls.VersionTLS10,
	}
	if s.TlsCA != "" {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(s.TlsCA)
		if err != nil {
			return nil, fmt.Errorf("Could not read CA certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return tlsConfig, nil
}
