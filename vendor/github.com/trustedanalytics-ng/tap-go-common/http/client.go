/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

const MaxIdleConnectionPerHost int = 20
const ConnectionTimeout = time.Duration(time.Second * 30)

var logger, _ = commonLogger.InitLogger("http")

func SetLoggerLevel(level string) error {
	return commonLogger.SetLoggerLevel(logger, level)
}

func GetHttpClientWithCertAndCa(certPem, keyPem, caPem string) (*http.Client, *http.Transport, error) {
	cert, ca, err := getCertKeyAndCa(certPem, keyPem, caPem)
	if err != nil {
		return nil, nil, err
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            ca,
		InsecureSkipVerify: IsInsecureSkipVerifyEnabled(),
		//ServerName: "kube-apiserver",  // if necessary, provide certificate name manually, after manual verification!
	}
	tlsConfig.BuildNameToCertificate()

	transport := prepareTransportWithProxy(tlsConfig, MaxIdleConnectionPerHost)

	client := &http.Client{Transport: transport, Timeout: ConnectionTimeout}

	return client, transport, nil
}

func GetHttpClientWithCertAndCaFromFile(certPemFile, keyPemFile, caPemFile string) (*http.Client, *http.Transport, error) {
	cert, err := tls.LoadX509KeyPair(certPemFile, keyPemFile)
	if err != nil {
		return nil, nil, err
	}

	caPemByte, err := ioutil.ReadFile(caPemFile)
	if err != nil {
		return nil, nil, err
	}

	ca := x509.NewCertPool()
	ca.AppendCertsFromPEM(caPemByte)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            ca,
		InsecureSkipVerify: IsInsecureSkipVerifyEnabled(),
	}
	tlsConfig.BuildNameToCertificate()

	transport := prepareTransportWithProxy(tlsConfig, MaxIdleConnectionPerHost)

	client := &http.Client{Transport: transport, Timeout: ConnectionTimeout}

	return client, transport, nil
}

func GetHttpClientWithCa(caPem string) (*http.Client, *http.Transport, error) {
	ca, err := getCa(caPem)
	if err != nil {
		return nil, nil, err
	}
	tlsConfig := &tls.Config{
		RootCAs:            ca,
		InsecureSkipVerify: IsInsecureSkipVerifyEnabled(),
	}
	tlsConfig.BuildNameToCertificate()

	transport := prepareTransportWithProxy(tlsConfig, MaxIdleConnectionPerHost)

	client := &http.Client{Transport: transport, Timeout: ConnectionTimeout}
	return client, transport, nil
}

func GetHttpClient() (*http.Client, *http.Transport, error) {
	return GetHttpClientWithCustomConnectionLimit(MaxIdleConnectionPerHost)
}

func GetHttpClientWithCustomSSLValidation(skipSSLValidation bool) (*http.Client, *http.Transport, error) {
	return GetHttpClientWithCustomConnectionLimitAndSSLValidation(MaxIdleConnectionPerHost, skipSSLValidation)
}

func GetHttpClientWithCustomConnectionLimit(maxIdleConnectionPerHost int) (*http.Client, *http.Transport, error) {
	return GetHttpClientWithCustomConnectionLimitAndSSLValidation(maxIdleConnectionPerHost, false)
}

func GetHttpClientWithCustomConnectionLimitAndSSLValidation(maxIdleConnectionPerHost int, skipSSLValidation bool) (*http.Client, *http.Transport, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: IsInsecureSkipVerifyEnabled() || skipSSLValidation,
	}
	tlsConfig.BuildNameToCertificate()

	transport := prepareTransportWithProxy(tlsConfig, maxIdleConnectionPerHost)

	client := &http.Client{Transport: transport, Timeout: ConnectionTimeout}
	return client, transport, nil
}

func IsInsecureSkipVerifyEnabled() bool {
	valueToParse := os.Getenv("INSECURE_SKIP_VERIFY")
	if valueToParse == "" {
		return false
	}
	insecureSkipVerify, err := strconv.ParseBool(valueToParse)
	if err != nil {
		logger.Panic("Can't read INSECURE_SKIP_VERIFY env!", err)
	}
	return insecureSkipVerify
}

func prepareTransportWithProxy(tlsConfig *tls.Config, maxIdleConnectionPerHost int) *http.Transport {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     tlsConfig,
		MaxIdleConnsPerHost: maxIdleConnectionPerHost,
		Dial: (&net.Dialer{ 
			Timeout: ConnectionTimeout,
		}).Dial,
	}
}

func getCertKeyAndCa(cert, key, ca string) (tls.Certificate, *x509.CertPool, error) {
	s_crt := strings.Replace(cert, " ", "\n", -1)
	s_crt = strings.Replace(s_crt, "-----BEGIN\nCERTIFICATE-----", "-----BEGIN CERTIFICATE-----", -1)
	s_crt = strings.Replace(s_crt, "-----END\nCERTIFICATE-----", "-----END CERTIFICATE-----", -1)

	s_key := strings.Replace(key, " ", "\n", -1)
	s_key = strings.Replace(s_key, "-----BEGIN\nRSA\nPRIVATE\nKEY-----", "-----BEGIN RSA PRIVATE KEY-----", -1)
	s_key = strings.Replace(s_key, "-----END\nRSA\nPRIVATE\nKEY-----", "-----END RSA PRIVATE KEY-----", -1)
	s_key = strings.Replace(s_key, "-----BEGIN\nPRIVATE\nKEY-----", "-----BEGIN PRIVATE KEY-----", -1)
	s_key = strings.Replace(s_key, "-----END\nPRIVATE\nKEY-----", "-----END PRIVATE KEY-----", -1)

	s_ca := strings.Replace(ca, " ", "\n", -1)
	s_ca = strings.Replace(s_ca, "-----BEGIN\nCERTIFICATE-----", "-----BEGIN CERTIFICATE-----", -1)
	s_ca = strings.Replace(s_ca, "-----END\nCERTIFICATE-----", "-----END CERTIFICATE-----", -1)

	certBytes := []byte(s_crt)
	keyBytes := []byte(s_key)
	caCert := []byte(s_ca)

	x509cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return x509cert, nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return x509cert, caCertPool, nil
}

func getCa(ca string) (*x509.CertPool, error) {
	s_ca := strings.Replace(ca, " ", "\n", -1)
	s_ca = strings.Replace(s_ca, "-----BEGIN\nCERTIFICATE-----", "-----BEGIN CERTIFICATE-----", -1)
	s_ca = strings.Replace(s_ca, "-----END\nCERTIFICATE-----", "-----END CERTIFICATE-----", -1)
	caCert := []byte(s_ca)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}
