package bux

import (
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_NewXpub will test the method NewXpub()
func (ts *EmbeddedDBTestSuite) TestClient_NewXpub() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)

			xPub2, err2 := tc.client.GetXpub(tc.ctx, testXPub)
			require.NoError(t, err2)
			assert.Equal(t, testXPubID, xPub2.ID)
		})

		ts.T().Run(testCase.name+" - valid with metadata (key->val)", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			opts := append(tc.client.DefaultModelOptions(), WithMetadata(testMetadataKey, testMetadataValue))

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, opts...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)
			assert.Equal(t, Metadata{testMetadataKey: testMetadataValue}, xPub.Metadata)

			xPub2, err2 := tc.client.GetXpub(tc.ctx, testXPub)
			require.NoError(t, err2)
			assert.Equal(t, testXPubID, xPub2.ID)
			assert.Equal(t, Metadata{testMetadataKey: testMetadataValue}, xPub2.Metadata)
		})

		ts.T().Run(testCase.name+" - valid with metadatas", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			opts := append(
				tc.client.DefaultModelOptions(),
				WithMetadatas(map[string]interface{}{
					testMetadataKey: testMetadataValue,
				}),
			)

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, opts...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)
			assert.Equal(t, Metadata{testMetadataKey: testMetadataValue}, xPub.Metadata)

			xPub2, err2 := tc.client.GetXpub(tc.ctx, testXPub)
			require.NoError(t, err2)
			assert.Equal(t, testXPubID, xPub2.ID)
			assert.Equal(t, Metadata{testMetadataKey: testMetadataValue}, xPub2.Metadata)
		})

		ts.T().Run(testCase.name+" - errors", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, err := tc.client.NewXpub(tc.ctx, "test", tc.client.DefaultModelOptions()...)
			assert.ErrorIs(t, err, utils.ErrXpubInvalidLength)

			_, err = tc.client.NewXpub(tc.ctx, "", tc.client.DefaultModelOptions()...)
			assert.ErrorIs(t, err, utils.ErrXpubInvalidLength)
		})

		ts.T().Run(testCase.name+" - duplicate xPub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)

			_, err2 := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			require.Error(t, err2)
		})
	}
}

// TestClient_GetXpub will test the method GetXpub()
func (ts *EmbeddedDBTestSuite) TestClient_GetXpub() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)

			xPub2, err2 := tc.client.GetXpub(tc.ctx, testXPub)
			require.NoError(t, err2)
			assert.Equal(t, testXPubID, xPub2.ID)

			xPub3, err3 := tc.client.GetXpubByID(tc.ctx, xPub2.ID)
			require.NoError(t, err3)
			assert.Equal(t, testXPubID, xPub3.ID)
		})

		ts.T().Run(testCase.name+" - error - invalid xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.GetXpub(tc.ctx, "test")
			require.Error(t, err)
			require.Nil(t, xPub)
			assert.ErrorIs(t, err, utils.ErrXpubInvalidLength)
		})

		ts.T().Run(testCase.name+" - error - missing xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.GetXpub(tc.ctx, testXPub)
			require.Error(t, err)
			require.Nil(t, xPub)
			assert.ErrorIs(t, err, ErrMissingXpub)
		})
	}
}

// TestClient_GetXpubByID will test the method GetXpubByID()
func (ts *EmbeddedDBTestSuite) TestClient_GetXpubByID() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			assert.Equal(t, testXPubID, xPub.ID)

			xPub2, err2 := tc.client.GetXpubByID(tc.ctx, xPub.ID)
			require.NoError(t, err2)
			assert.Equal(t, testXPubID, xPub2.ID)
		})

		ts.T().Run(testCase.name+" - error - invalid xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.GetXpubByID(tc.ctx, "test")
			require.Error(t, err)
			require.Nil(t, xPub)
		})

		ts.T().Run(testCase.name+" - error - missing xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			xPub, err := tc.client.GetXpubByID(tc.ctx, testXPub)
			require.Error(t, err)
			require.Nil(t, xPub)
			assert.ErrorIs(t, err, ErrMissingXpub)
		})
	}
}
