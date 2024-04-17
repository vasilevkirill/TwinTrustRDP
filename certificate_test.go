package main

import (
	"os"
	"testing"
)

func TestGenerateCertificate(t *testing.T) {

	// Установка временных путей для файлов сертификата и закрытого ключа.
	certFileKey = "./testdata/cert.key"
	certFileCer = "./testdata/cert.cer"
	defer func() {
		err := os.RemoveAll("./testdata")
		if err != nil {
			t.Errorf("Remove folder %s error: %s", "./testdata", err.Error())
		}
	}()

	err := generateCertificate()
	if err != nil {
		t.Errorf("generateCertificate() error = %v, want nil", err)
	}

	// Проверка существования файлов сертификата и закрытого ключа.
	if _, err := os.Stat(certFileKey); os.IsNotExist(err) {
		t.Errorf("generateCertificate() failed to create private key file")
	}
	if _, err := os.Stat(certFileCer); os.IsNotExist(err) {
		t.Errorf("generateCertificate() failed to create certificate file")
	}
}
