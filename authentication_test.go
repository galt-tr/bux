package bux

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// testAccessKey         = "9b2a4421edd88782a193ea8195cce1fe9b632df575c88d70f20a1fdf6835b764"
	// testAccessKeyAddress  = "1HuoHijPa7BqQNiV953pd3taqnmyhgDXFt"
	// testAccessKeyID       = "b7b91e8aca22b4ee33f3f0e48c00cd4631dc2dbba1f773829883eaae42fa2234"
	// testAccessKeyPKH      = "b97e4834a13d188ab0588dc2aaff11a6658771cd"
	// testAccessKeyPublic   = "02719a5e3623bee13f8116f1db4ee54603c993e020087960f31d2e0b4cbd97d175"
	// testSignatureAuthNonce = `dec0535f13b7ed61c2b188b7fe8fd5f578d6931aa90b6063c653ce0f8eefacf1`
	// testSignatureAuthTime  = "1643828414038"
	// testSignatureXpub     = `xpub661MyMwAqRbcFnj7dmEoX4ULYMJ2vxFBkH3oGrpuQMHTMpxUEGND1UXwskzgtUj6R7i9dRNGYj6NYuXWKVM5yAJYjSGuvBJfDTpqjsh8a3T`
	testBodyContents      = `{"test_field":"test_value"}`
	testSignature         = `HxNguR72c6BV7tKNn5BQ3/mS2+RX3BGyQHFfVfQ3v4mVdAuh+w32QsFYxsB13KiXuRJ7ZnN7C8RhkAtLi/qvH88=`
	testSignatureAuthHash = `5858adf09a0cc01f6d3a4d377f010408313031bb96b40d98e6edccf18c26464e`
	testXpub              = "xpub661MyMwAqRbcH3WGvLjupmr43L1GVH3MP2WQWvdreDraBeFJy64Xxv4LLX9ZVWWz3ZjZkMuZtSsc9qH9JZR74bR4PWkmtEvP423r6DJR8kA"
	testXpubHash          = "d8c2bed524071d72d859caf90da5f448b5861cd4d4fd47697f94166c13c5a987"
)

// TestClient_AuthenticateRequest will test the method AuthenticateRequest()
func TestClient_AuthenticateRequest(t *testing.T) {
	t.Parallel()

	t.Run("valid xpub", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, testXpub)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{""}, false, false, false,
		)
		require.NoError(t, err)
		require.NotNil(t, req)

		// Test the request
		x, ok := GetXpubFromRequest(req)
		assert.Equal(t, testXpub, x)
		assert.Equal(t, true, ok)

		x, ok = GetXpubHashFromRequest(req)
		assert.Equal(t, testXpubHash, x)
		assert.Equal(t, true, ok)
	})

	t.Run("error - admin required", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, testXpub)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{""}, true, false, false,
		)
		require.Error(t, err)
		require.NotNil(t, req)
		assert.ErrorIs(t, err, ErrNotAdminKey)
	})

	t.Run("error - admin key - missing body", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, testXpub)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{testXpub}, true, false, false,
		)
		require.Error(t, err)
		require.NotNil(t, req)
		assert.ErrorIs(t, err, ErrMissingBody)
	})

	t.Run("error - admin key - missing signature", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", bytes.NewReader([]byte(`{}`)))
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, testXpub)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{testXpub}, true, false, false,
		)
		require.Error(t, err)
		require.NotNil(t, req)
		assert.ErrorIs(t, err, ErrMissingSignature)
	})

	t.Run("admin key - valid signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var req *http.Request
		req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "", bytes.NewReader([]byte(`{}`)))
		require.NoError(t, err)
		require.NotNil(t, req)

		var authData *AuthPayload
		authData, err = createSignature(key, `{}`)
		require.NoError(t, err)
		require.NotNil(t, authData)

		err = SetSignature(&req.Header, key, `{}`)
		require.NoError(t, err)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{authData.xPub}, true, false, false,
		)
		require.NoError(t, err)
		require.NotNil(t, req)
	})

	t.Run("admin key - signing disabled", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", bytes.NewReader([]byte(`{}`)))
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, testXpub)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{testXpub}, true, false, true,
		)
		require.NoError(t, err)
		require.NotNil(t, req)
	})

	t.Run("no authentication header set", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{""}, false, false, false,
		)
		require.Error(t, err)
		require.NotNil(t, req)

		// Test the request
		x, ok := GetXpubFromRequest(req)
		assert.Equal(t, "", x)
		assert.Equal(t, false, ok)

		x, ok = GetXpubHashFromRequest(req)
		assert.Equal(t, "", x)
		assert.Equal(t, false, ok)
	})

	t.Run("invalid xpub length", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set(AuthHeader, "invalid-length")

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		req, err = client.AuthenticateRequest(
			context.Background(), req, []string{""}, false, false, false,
		)
		require.Error(t, err)
		require.NotNil(t, req)

		// Test the request
		x, ok := GetXpubFromRequest(req)
		assert.Equal(t, "", x)
		assert.Equal(t, false, ok)

		x, ok = GetXpubHashFromRequest(req)
		assert.Equal(t, "", x)
		assert.Equal(t, false, ok)
	})
}

// Test_verifyKeyXPub will test the method verifyKeyXPub()
func Test_verifyKeyXPub(t *testing.T) {
	t.Parallel()

	t.Run("error - missing auth data", func(t *testing.T) {

		err := verifyKeyXPub(testXpub, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingSignature)
	})

	t.Run("error - missing auth signature", func(t *testing.T) {
		err := checkSignatureRequirements(&AuthPayload{})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingSignature)
	})

	t.Run("error - auth hash mismatch", func(t *testing.T) {
		err := checkSignatureRequirements(&AuthPayload{
			AuthHash:     "bad-hash",
			BodyContents: testBodyContents,
			Signature:    testSignature,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAuhHashMismatch)
	})

	t.Run("error - signature expired", func(t *testing.T) {
		err := checkSignatureRequirements(&AuthPayload{
			AuthHash:     testSignatureAuthHash,
			BodyContents: testBodyContents,
			Signature:    testSignature,
			AuthTime:     1643828414038,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSignatureExpired)
	})

	t.Run("error - bad xpub", func(t *testing.T) {
		err := verifyKeyXPub("invalid-key", &AuthPayload{
			AuthHash:     testSignatureAuthHash,
			BodyContents: testBodyContents,
			Signature:    testSignature,
			AuthTime:     time.Now().UnixMilli(),
		})
		require.Error(t, err)
	})

	t.Run("error - invalid signature - time is wrong", func(t *testing.T) {
		err := checkSignatureRequirements(&AuthPayload{
			AuthHash:     testSignatureAuthHash,
			BodyContents: testBodyContents,
			Signature:    testSignature,
			AuthTime:     0,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSignatureExpired)
	})

	t.Run("valid signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		authData, err2 := createSignature(key, testBodyContents)
		require.NoError(t, err2)
		require.NotNil(t, authData)

		err = verifyKeyXPub(authData.xPub, &AuthPayload{
			AuthHash:     authData.AuthHash,
			AuthNonce:    authData.AuthNonce,
			AuthTime:     authData.AuthTime,
			BodyContents: testBodyContents,
			Signature:    authData.Signature,
		})
		require.NoError(t, err)
	})
}

// TestCreateSignature will test the method CreateSignature()
func TestCreateSignature(t *testing.T) {
	t.Parallel()

	t.Run("valid signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var sig string
		sig, err = CreateSignature(key, testBodyContents)
		require.NoError(t, err)
		require.NotNil(t, sig)
		assert.Greater(t, len(sig), 40)
	})

	t.Run("missing key", func(t *testing.T) {
		sig, err := CreateSignature(nil, testBodyContents)
		require.Error(t, err)
		assert.Equal(t, "", sig)
	})

	t.Run("missing body contents - still has signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var sig string
		sig, err = CreateSignature(key, "")
		require.NoError(t, err)
		require.NotNil(t, sig)
		assert.Greater(t, len(sig), 40)
	})
}

// Test_createSignature will test the method createSignature()
func Test_createSignature(t *testing.T) {
	t.Parallel()

	t.Run("valid signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var authData *AuthPayload
		authData, err = createSignature(key, testBodyContents)
		require.NoError(t, err)
		require.NotNil(t, authData)

		assert.Equal(t, 111, len(authData.xPub))
		assert.Equal(t, 64, len(authData.AuthHash))
		assert.Equal(t, 64, len(authData.AuthNonce))
		assert.Greater(t, authData.AuthTime, time.Now().Add(-1*time.Second).UnixMilli())

		err = verifyKeyXPub(authData.xPub, &AuthPayload{
			AuthHash:     authData.AuthHash,
			AuthNonce:    authData.AuthNonce,
			AuthTime:     authData.AuthTime,
			BodyContents: testBodyContents,
			Signature:    authData.Signature,
		})
		require.NoError(t, err)
	})

	t.Run("error - missing key", func(t *testing.T) {
		authData, err := createSignature(nil, testBodyContents)
		require.Error(t, err)
		require.Nil(t, authData)
	})

	t.Run("error - empty body - valid signature", func(t *testing.T) {
		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var authData *AuthPayload
		authData, err = createSignature(key, "")
		require.NoError(t, err)
		require.NotNil(t, authData)

		assert.Equal(t, 111, len(authData.xPub))
		assert.Equal(t, 0, len(authData.AuthHash))
		assert.Equal(t, 64, len(authData.AuthNonce))
		assert.Greater(t, authData.AuthTime, time.Now().Add(-1*time.Second).UnixMilli())

		err = verifyKeyXPub(authData.xPub, &AuthPayload{
			AuthHash:     authData.AuthHash,
			AuthNonce:    authData.AuthNonce,
			AuthTime:     authData.AuthTime,
			BodyContents: "",
			Signature:    authData.Signature,
		})
		require.NoError(t, err)
	})
}

// TestSetSignature will test the method SetSignature()
func TestSetSignature(t *testing.T) {
	t.Parallel()

	t.Run("error - bad signature", func(t *testing.T) {
		err := SetSignature(nil, nil, testBodyContents)
		require.Error(t, err)
	})

	t.Run("valid set headers", func(t *testing.T) {
		emptyHeaders := &http.Header{}

		key, err := bitcoin.GenerateHDKey(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		require.NotNil(t, key)

		var xPub string
		xPub, err = bitcoin.GetExtendedPublicKey(key)
		require.NoError(t, err)
		require.NotEmpty(t, xPub)

		err = SetSignature(emptyHeaders, key, testBodyContents)
		require.NoError(t, err)

		assert.NotEmpty(t, emptyHeaders.Get(AuthHeader))
		assert.NotEmpty(t, emptyHeaders.Get(AuthHeaderHash))
		assert.NotEmpty(t, emptyHeaders.Get(AuthHeaderNonce))
		assert.NotEmpty(t, emptyHeaders.Get(AuthHeaderTime))
		assert.NotEmpty(t, emptyHeaders.Get(AuthSignature))

		authTime, _ := strconv.Atoi(emptyHeaders.Get(AuthHeaderTime))
		err = verifyKeyXPub(xPub, &AuthPayload{
			AuthHash:     emptyHeaders.Get(AuthHeaderHash),
			AuthNonce:    emptyHeaders.Get(AuthHeaderNonce),
			AuthTime:     int64(authTime),
			BodyContents: testBodyContents,
			Signature:    emptyHeaders.Get(AuthSignature),
		})
		require.NoError(t, err)
	})
}

// Test_getSigningMessage will test the method Test_getSigningMessage()
func Test_getSigningMessage(t *testing.T) {
	t.Parallel()

	t.Run("valid format", func(t *testing.T) {
		message := getSigningMessage(testXpub, &AuthPayload{
			AuthHash:  testXpubHash,
			AuthNonce: "auth-nonce",
			AuthTime:  12345678,
		})
		assert.Equal(t, fmt.Sprintf("%s%s%s%d", testXpub, testXpubHash, "auth-nonce", 12345678), message)
	})
}

// TestGetXpubFromRequest will test the method GetXpubFromRequest()
func TestGetXpubFromRequest(t *testing.T) {
	t.Parallel()

	t.Run("valid value", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = setOnRequest(req, xPubKey, testXpub)

		xPub, success := GetXpubFromRequest(req)
		assert.Equal(t, testXpub, xPub)
		assert.Equal(t, true, success)
	})

	t.Run("no value", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		xPub, success := GetXpubFromRequest(req)
		assert.Equal(t, "", xPub)
		assert.Equal(t, false, success)
	})
}

// TestGetXpubHashFromRequest will test the method GetXpubHashFromRequest()
func TestGetXpubHashFromRequest(t *testing.T) {
	t.Parallel()

	t.Run("valid value", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = setOnRequest(req, xPubHashKey, testXpubHash)

		xPubHash, success := GetXpubHashFromRequest(req)
		assert.Equal(t, testXpubHash, xPubHash)
		assert.Equal(t, true, success)
	})

	t.Run("no value", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		xPubHash, success := GetXpubHashFromRequest(req)
		assert.Equal(t, "", xPubHash)
		assert.Equal(t, false, success)
	})
}
