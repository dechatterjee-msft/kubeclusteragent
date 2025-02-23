package calico

import (
	"bytes"
	"context"
	"fmt"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/osutility/linux"
	"text/template"
)

var ipReservationTemplate = `apiVersion: crd.projectcalico.org/v1
kind: IPReservation
metadata:
  name: snc-pod-cidr-reserved-pool
spec:
  reservedCIDRs:
    - {{ .PodIPv4 }}
    - {{ .PodIPv6 }}`

type ipReservation struct {
	PodIPv4 string
	PodIPv6 string
}

func ConfigurePodIPReservation(ctx context.Context, ou linux.OSUtil) (string, error) {
	calicoIPReservation := &ipReservation{
		PodIPv4: constants.PodIPv4Reservation,
		PodIPv6: constants.PodIPv6Reservation,
	}
	tmpl, err := template.New("calicoReservation").Parse(ipReservationTemplate)
	if err != nil {
		return "", fmt.Errorf("parse kubeadm configuration template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, calicoIPReservation); err != nil {
		return "", fmt.Errorf("execute template for pod ip reservation: %w", err)
	}
	var calicoIPReservationFile = "/tmp/calico_ip_reservation.yaml"
	err = ou.Filesystem().WriteFile(ctx, calicoIPReservationFile, buf.Bytes(), constants.FilePerm)
	if err != nil {
		return "", err
	}
	response, err := ou.Kubectl().RunWithResponse(ctx, "apply", "-f", calicoIPReservationFile)
	if err != nil {
		return "", err
	}
	return response, nil
}
