// Copyright 2012 Aaron Jacobs. All Rights Reserved.
// Author: aaronjjacobs@gmail.com (Aaron Jacobs)
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

package siv_test

import (
	"crypto/rand"
	"testing"

	. "github.com/jacobsa/oglematchers"
	. "github.com/jacobsa/ogletest"
	"github.com/yueryou/jacobsa-crypto/siv"
	aes_testing "github.com/yueryou/jacobsa-crypto/testing"
)

func TestEncrypt(t *testing.T) { RunTests(t) }

////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////

type EncryptTest struct{}

func init() { RegisterTestSuite(&EncryptTest{}) }

////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////

func (t *EncryptTest) NilKey() {
	key := []byte(nil)
	plaintext := []byte{}

	_, err := siv.Encrypt(nil, key, plaintext, nil)
	ExpectThat(err, Error(HasSubstr("-byte")))
}

func (t *EncryptTest) ShortKey() {
	key := make([]byte, 31)
	plaintext := []byte{}

	_, err := siv.Encrypt(nil, key, plaintext, nil)
	ExpectThat(err, Error(HasSubstr("-byte")))
}

func (t *EncryptTest) LongKey() {
	key := make([]byte, 65)
	plaintext := []byte{}

	_, err := siv.Encrypt(nil, key, plaintext, nil)
	ExpectThat(err, Error(HasSubstr("-byte")))
}

func (t *EncryptTest) TooMuchAssociatedData() {
	key := make([]byte, 64)
	plaintext := []byte{}
	associated := make([][]byte, 127)

	_, err := siv.Encrypt(nil, key, plaintext, associated)
	ExpectThat(err, Error(HasSubstr("associated")))
	ExpectThat(err, Error(HasSubstr("126")))
}

func (t *EncryptTest) JustLittleEnoughAssociatedData() {
	key := make([]byte, 64)
	plaintext := []byte{}
	associated := make([][]byte, 126)

	_, err := siv.Encrypt(nil, key, plaintext, associated)
	ExpectEq(nil, err)
}

func (t *EncryptTest) DoesntClobberAssociatedSlice() {
	key := make([]byte, 32)
	plaintext := []byte{}

	associated0 := aes_testing.FromRfcHex("deadbeef")
	associated1 := aes_testing.FromRfcHex("feedface")
	associated2 := aes_testing.FromRfcHex("ba5eba11")
	associated := [][]byte{associated0, associated1, associated2}

	// Call with a slice of associated missing the last element. The last element
	// shouldn't be clobbered.
	_, err := siv.Encrypt(nil, key, plaintext, associated[:2])
	AssertEq(nil, err)

	ExpectThat(
		associated,
		ElementsAre(
			DeepEquals(associated0),
			DeepEquals(associated1),
			DeepEquals(associated2),
		))
}

func (t *EncryptTest) OutputIsDeterministic() {
	cases := aes_testing.EncryptCases()
	AssertGt(len(cases), 37)
	c := cases[37]

	output0, err := siv.Encrypt(nil, c.Key, c.Plaintext, c.Associated)
	AssertEq(nil, err)

	output1, err := siv.Encrypt(nil, c.Key, c.Plaintext, c.Associated)
	AssertEq(nil, err)

	output2, err := siv.Encrypt(nil, c.Key, c.Plaintext, c.Associated)
	AssertEq(nil, err)

	AssertGt(len(output0), 0)
	ExpectThat(output0, DeepEquals(output1))
	ExpectThat(output0, DeepEquals(output2))
}

func (t *EncryptTest) Rfc5297TestCaseA1() {
	key := aes_testing.FromRfcHex(
		"fffefdfc fbfaf9f8 f7f6f5f4 f3f2f1f0" +
			"f0f1f2f3 f4f5f6f7 f8f9fafb fcfdfeff")

	plaintext := aes_testing.FromRfcHex(
		"11223344 55667788 99aabbcc ddee")

	associated := [][]byte{
		aes_testing.FromRfcHex(
			"10111213 14151617 18191a1b 1c1d1e1f" +
				"20212223 24252627"),
	}

	expected := aes_testing.FromRfcHex(
		"85632d07 c6e8f37f 950acd32 0a2ecc93" +
			"40c02b96 90c4dc04 daef7f6a fe5c")

	output, err := siv.Encrypt(nil, key, plaintext, associated)
	AssertEq(nil, err)
	ExpectThat(output, DeepEquals(expected))
}

func (t *EncryptTest) Rfc5297TestCaseA2() {
	key := aes_testing.FromRfcHex(
		"7f7e7d7c 7b7a7978 77767574 73727170" +
			"40414243 44454647 48494a4b 4c4d4e4f")

	plaintext := aes_testing.FromRfcHex(
		"74686973 20697320 736f6d65 20706c61" +
			"696e7465 78742074 6f20656e 63727970" +
			"74207573 696e6720 5349562d 414553")

	associated := [][]byte{
		aes_testing.FromRfcHex(
			"00112233 44556677 8899aabb ccddeeff" +
				"deaddada deaddada ffeeddcc bbaa9988" +
				"77665544 33221100"),
		aes_testing.FromRfcHex(
			"10203040 50607080 90a0"),
		aes_testing.FromRfcHex(
			"09f91102 9d74e35b d84156c5 635688c0"),
	}

	expected := aes_testing.FromRfcHex(
		"7bdb6e3b 432667eb 06f4d14b ff2fbd0f" +
			"cb900f2f ddbe4043 26601965 c889bf17" +
			"dba77ceb 094fa663 b7a3f748 ba8af829" +
			"ea64ad54 4a272e9c 485b62a3 fd5c0d")

	output, err := siv.Encrypt(nil, key, plaintext, associated)
	AssertEq(nil, err)
	ExpectThat(output, DeepEquals(expected))
}

func (t *EncryptTest) GeneratedTestCases() {
	cases := aes_testing.EncryptCases()
	AssertGe(len(cases), 100)

	for i, c := range cases {
		output, err := siv.Encrypt(nil, c.Key, c.Plaintext, c.Associated)
		AssertEq(nil, err, "Case %d: %v", i, c)
		ExpectThat(output, DeepEquals(c.Output), "Case %d: %v", i, c)
	}
}

////////////////////////////////////////////////////////////////////////
// Benchmarks
////////////////////////////////////////////////////////////////////////

func benchmarkEncrypt(
	b *testing.B,
	size int) {
	var err error

	// Generate the appropriate amount of random data.
	plaintext := make([]byte, size)
	_, err = rand.Read(plaintext)
	if err != nil {
		b.Fatalf("rand.Read: %v", err)
	}

	// Create a random key.
	const keyLen = 32
	key := make([]byte, keyLen)

	_, err = rand.Read(key)
	if err != nil {
		b.Fatalf("rand.Read: %v", err)
	}

	// Repeatedly encrypt.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = siv.Encrypt(nil, key, plaintext, nil)
		if err != nil {
			b.Fatalf("Encrypt: %v", err)
		}
	}

	b.SetBytes(int64(size))
}

func BenchmarkEncrypt1KiB(b *testing.B) {
	const size = 1 << 10
	benchmarkEncrypt(b, size)
}

func BenchmarkEncrypt1MiB(b *testing.B) {
	const size = 1 << 20
	benchmarkEncrypt(b, size)
}

func BenchmarkEncrypt16MiB(b *testing.B) {
	const size = 1 << 24
	benchmarkEncrypt(b, size)
}
