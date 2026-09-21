package crypto

import "testing"

func TestPassword_HashPassword(t *testing.T) {

	realPassword := "rahasia123"

	hashed,err := HashPassword(realPassword)

	if err != nil {
		t.Fatalf("failed to hash password:%v",err)
	}

	if len(hashed) == 0 {
		t.Fatalf("hash length : %d",len(hashed))
	}

}

func TestPassword_CheckPassword(t *testing.T) {

	realPassword := "rahasia123"

	hashed,err := HashPassword(realPassword)

	if err != nil {
		t.Fatalf("failed to hash password:%v",err)
	}

	checked := CheckPassword(realPassword,hashed)

	if checked != true {
		t.Fatalf("failed to compare password:%v",checked)
	}

}

func TestPassword_CheckPassword_WrongPassword(t *testing.T) {
    realPassword := "rahasia123"

    hashed, err := HashPassword(realPassword)
    if err != nil {
        t.Fatalf("failed to hash password: %v", err)
    }

    checked := CheckPassword("password-salah", hashed)

    if checked {
        t.Fatalf("expected password check to fail")
    }
}