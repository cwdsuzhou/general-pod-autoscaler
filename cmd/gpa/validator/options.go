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
	"fmt"
	"net"

	"github.com/spf13/pflag"
)

var (
	Version = "unknown"
)

type ServerRunOptions struct {
	Address              string
	Port                 int
	TlsCA                string
	TlsCert              string
	TlsKey               string
	IgnoreLabelKeys      string
	ShowVersion          bool
	SrcResourceName      string
	DstResourceName      string
	AllowDescheduleCount int
}

func NewServerRunOptions() *ServerRunOptions {
	options := &ServerRunOptions{}
	options.addFlags()
	return options
}

func (s *ServerRunOptions) addFlags() {
	pflag.StringVar(&s.Address, "address", "0.0.0.0", "The address of scheduler manager.")
	pflag.IntVar(&s.Port, "port", 8080, "The port of scheduler manager.")
	pflag.StringVar(&s.TlsCert, "tlscert", "", "Path to TLS certificate file")
	pflag.StringVar(&s.TlsKey, "tlskey", "", "Path to TLS key file")
	pflag.StringVar(&s.TlsCA, "CA", "", "Path to certificate file")
	pflag.BoolVar(&s.ShowVersion, "version", false, "Show version.")
}

func (s *ServerRunOptions) Validate() error {
	address := net.ParseIP(s.Address)
	if address.To4() == nil {
		return fmt.Errorf("%v is not a valid IP address\n", s.Address)
	}
	return nil
}
