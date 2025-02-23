package linux

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var x = `[check-expiration] Reading configuration from the cluster...
[check-expiration] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'

CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
admin.conf                 Dec 13, 2024 09:12 UTC   363d            ca                      no      
apiserver                  Dec 13, 2024 09:12 UTC   363d            ca                      no      
apiserver-etcd-client      Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
apiserver-kubelet-client   Dec 13, 2024 09:12 UTC   363d            ca                      no      
controller-manager.conf    Dec 13, 2024 09:12 UTC   363d            ca                      no      
etcd-healthcheck-client    Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
etcd-peer                  Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
etcd-server                Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
front-proxy-client         Dec 13, 2024 09:12 UTC   363d            front-proxy-ca          no      
scheduler.conf             Dec 13, 2024 09:12 UTC   363d            ca                      no      

CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
ca                      Dec 10, 2033 22:33 UTC   9y              no      
etcd-ca                 Dec 10, 2033 22:33 UTC   9y              no      
front-proxy-ca          Dec 10, 2033 22:33 UTC   9y              no      
`

func Test_evaluateOverallCertsExpiration(t *testing.T) {
	type args struct {
		expiryInfo string
	}
	testMap := make(map[string]int64)
	testMap["admin.conf"] = 363
	testMap["apiserver"] = 363
	testMap["apiserver-etcd-client"] = 363
	testMap["apiserver-kubelet-client"] = 363
	testMap["controller-manager.conf"] = 363
	testMap["etcd-healthcheck-client"] = 363
	testMap["etcd-peer"] = 363
	testMap["etcd-server"] = 363
	testMap["front-proxy-client"] = 363
	testMap["scheduler.conf"] = 363
	tests := []struct {
		name    string
		args    args
		want    int
		want1   map[string]int64
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "AllSameExpiryDate", args: struct{ expiryInfo string }{expiryInfo: x}, want: 363, want1: testMap, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := evaluateOverallCertsExpiration(tt.args.expiryInfo)
			if err != nil {
				t.Fail()
			}
			assert.Equalf(t, tt.want, got, "evaluateOverallCertsExpiration(%v)", tt.args.expiryInfo)
			assert.Equalf(t, tt.want1, got1, "evaluateOverallCertsExpiration(%v)", tt.args.expiryInfo)
		})
	}
}

var y = `[check-expiration] Reading configuration from the cluster...
[check-expiration] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'

CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
admin.conf                 Dec 13, 2024 09:12 UTC   363d            ca                      no      
apiserver                  Dec 13, 2024 09:12 UTC   363d            ca                      no      
apiserver-etcd-client      Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
apiserver-kubelet-client   Dec 13, 2024 09:12 UTC   363d            ca                      no      
controller-manager.conf    Dec 13, 2024 09:12 UTC   363d            ca                      no      
etcd-healthcheck-client    Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
etcd-peer                  Dec 13, 2024 09:12 UTC   363d            etcd-ca                 no      
etcd-server                Dec 13, 2024 09:12 UTC   250d            etcd-ca                 no      
front-proxy-client         Dec 13, 2024 09:12 UTC   363d            front-proxy-ca          no      
scheduler.conf             Dec 13, 2024 09:12 UTC   363d            ca                      no      

CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
ca                      Dec 10, 2033 22:33 UTC   9y              no      
etcd-ca                 Dec 10, 2033 22:33 UTC   9y              no      
front-proxy-ca          Dec 10, 2033 22:33 UTC   9y              no      
`

func Test_evaluateOverallCertsExpiration_AllNotRotated(t *testing.T) {
	type args struct {
		expiryInfo string
	}
	testMap := make(map[string]int64)
	testMap["admin.conf"] = 363
	testMap["apiserver"] = 363
	testMap["apiserver-etcd-client"] = 363
	testMap["apiserver-kubelet-client"] = 363
	testMap["controller-manager.conf"] = 363
	testMap["etcd-healthcheck-client"] = 363
	testMap["etcd-peer"] = 363
	testMap["etcd-server"] = 250
	testMap["front-proxy-client"] = 363
	testMap["scheduler.conf"] = 363
	tests := []struct {
		name    string
		args    args
		want    int
		want1   map[string]int64
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "DifferentExpiryDate", args: struct{ expiryInfo string }{expiryInfo: y}, want: 250, want1: testMap, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := evaluateOverallCertsExpiration(tt.args.expiryInfo)
			if err != nil {
				t.Fail()
			}
			assert.Equalf(t, tt.want, got, "evaluateOverallCertsExpiration(%v)", tt.args.expiryInfo)
			assert.Equalf(t, tt.want1, got1, "evaluateOverallCertsExpiration(%v)", tt.args.expiryInfo)
		})
	}
}
