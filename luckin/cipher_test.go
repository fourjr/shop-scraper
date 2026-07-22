package luckin

import (
	"log"
	"testing"
)

const str = "LxtDNh2yCQlLTKPxD3N_4uV8UfRQEoukmnSMKu0-7T4_sSrEGEw4Mg2JzSZ6-kEYUkdWtaxpxNVFUuQo9z2GsAo4XrIeXBoaOsvq_yJ8HShLMf1yqCTU8AJEKQYGBqTBARsBjdJOysHUggfjt5wPDiRNqqz0qdZspb2MmYtHNC-fsB8gX21g9SqlbflhuwBO70hIwV1zfJbsfr-MSBts-bQG0hTb9hwc4DOLtgC1EpfkZCPAJs0X0XNLLHo0dEqmfufmen-FtgAHGUCDUOl6pnWYtcmmr0MgrpDcssswa9bZSL_msAH_-tGYAoixjC3ff31KWkx0-iizq7T1f20LqBEn4ZyVs0uWoQ2npH8RGc86WDtHfGao6vLm-doCj2yOTNAeJNpXAY1WSn_NuwSJDloH-7n1YTrF92uxor2h2J6CjNIOMFwFDyofKo2fzwKOsPuCVYqym53mMBZWQ1vsKq1xcpc7fVUOz-nEhAaoye4Kab_anlEqC54aQVaeQZt8D6Y5Oaz43hHbhxi__Viw45R_38D7NQpd1MSEznMdZqwDy7IlRl4QpujYwCOoLV2BXgNgiJNsQKQVfMhryi_zW9sVM9B_nVgVX7nCdiVueKv679T7durY7Wv2f2GcBcxgQ_WO1FcDmTppllBArlUpRpF92yHVilkc1FVwVu1IbF8xajSooR_1gPaGqD3hJErvRwFfoD45RhgTd_ogfPoZmeCn_rIoO5I1bi1u1gA0wvKmIffoeov-Oqjxm5J16NArtDlN4NNkw6T0Whsi4aahm5Cj0Dfq6Zp4yCjc-FNRmgfrgpqpsdqX1hQpqJp4R3LCRdN5cVzQw3YTPwHO5A0yOLFGU8ZPh_BxRKpYIGzwm5lxa24iF6vapcdZVaOtT1N4DEXMM9kItiUk36fGGdvbJPna0uwH_WvcBXE5JoESJ37oNcmPDoocmjj5iAA5ENfVUeb5D5OOVg1iOcIFR8E_v1_l8JLUmUgsKdWEOh3chLP6N8CvLljppdvjXWp25-NRPynsbwIEa89HyCThSzbfasgFmOwQV93_KjpN7j1AJvKQgLTmldhM6Ji749TQN9iTongwgckVW5WRZWODO6tG-5d-zgYFfHFYGcKe5gwL8eQ="

func TestLength(t *testing.T) {
	log.Printf("Length of string: %d", len(str))
}

// func TestDecrypt(t *testing.T) {
// 	plaintext, err := Decrypt(str)
// 	if err != nil {
// 		t.Fatalf("Decryption failed: %v", err)
// 	}
// 	log.Printf("Decrypted string: %s", plaintext)

// 	reencrypted, err := Encrypt(plaintext)
// 	if err != nil {
// 		t.Fatalf("Re-encryption failed: %v", err)
// 	}
// 	if reencrypted != str {
// 		t.Fatal("Re-encrypted string does not match the original ciphertext")
// 	}
// }

// func TestEncrypt(t *testing.T) {
// 	const plaintext = "hello, world"
// 	const expected = "GLnaJjFBDFoMoqFiMMyonw=="

// 	encrypted, err := Encrypt(plaintext)
// 	if err != nil {
// 		t.Fatalf("Encryption failed: %v", err)
// 	}
// 	if encrypted != expected {
// 		t.Fatalf("Encrypt(%q) = %q, want %q", plaintext, encrypted, expected)
// 	}
// }

// func TestEncryptDecryptRoundTrip(t *testing.T) {
// 	testCases := []string{
// 		"",
// 		"short",
// 		"exactly16bytes!!",
// 		"longer than one AES block",
// 		"茶瑞幸 ☕",
// 	}

// 	for _, plaintext := range testCases {
// 		t.Run(plaintext, func(t *testing.T) {
// 			encrypted, err := Encrypt(plaintext)
// 			if err != nil {
// 				t.Fatalf("Encryption failed: %v", err)
// 			}

// 			decrypted, err := Decrypt(encrypted)
// 			if err != nil {
// 				t.Fatalf("Decryption failed: %v", err)
// 			}
// 			if decrypted != plaintext {
// 				t.Fatalf("Decrypt(Encrypt(%q)) = %q", plaintext, decrypted)
// 			}
// 		})
// 	}
// }

func TestRequestHtml(t *testing.T) {
	key, err := getEncryptKey()
	if err != nil {
		t.Fatalf("RequestHtml failed: %v", err)
	}
	log.Printf("Encrypt key: %s", key)
}
