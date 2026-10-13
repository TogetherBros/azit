package login

import "testing"

func TestNewPKCE(t *testing.T) {
	verifier, challenge, err := newPKCE()
	if err != nil {
		t.Fatal(err)
	}

	// 32 바이트 → 패딩 없는 base64 는 43 글자 (RFC 7636 허용 범위 43~128)
	if len(verifier) != 43 {
		t.Errorf("verifier 길이 = %d, 43 이어야 해요", len(verifier))
	}

	// RFC 7636 부록 B 예시값으로 해시 방식 확인
	const rfcVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := challengeOf(rfcVerifier); got != rfcChallenge {
		t.Errorf("challengeOf = %s, want %s", got, rfcChallenge)
	}

	if challenge != challengeOf(verifier) {
		t.Error("challenge 가 verifier 해시와 달라요")
	}
}

func TestNewPKCEUnique(t *testing.T) {
	a, _, _ := newPKCE()
	b, _, _ := newPKCE()
	if a == b {
		t.Error("verifier 가 두 번 같게 나왔어요")
	}
}
