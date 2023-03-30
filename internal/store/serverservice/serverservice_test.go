package serverservice

import (
	"context"
	"encoding/json"
	"net"

	// _ "net/http/pprof"

	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/alloy/internal/fixtures"
	"github.com/metal-toolbox/alloy/internal/model"
	"github.com/stretchr/testify/assert"
	serverserviceapi "go.hollow.sh/serverservice/pkg/api/v1"
)

func Test_validateRequiredAttribtues(t *testing.T) {
	// nolint:govet // ignore struct alignment in test
	cases := []struct {
		name              string
		server            *serverserviceapi.Server
		secret            *serverserviceapi.ServerCredential
		expectCredentials bool
		expectedErr       string
	}{
		{
			"server object nil",
			nil,
			nil,
			true,
			"server object nil",
		},
		{
			"server credential object nil",
			&serverserviceapi.Server{},
			nil,
			true,
			"server credential object nil",
		},
		{
			"server attributes slice empty",
			&serverserviceapi.Server{},
			&serverserviceapi.ServerCredential{},
			true,
			"server attributes slice empty",
		},
		{
			"BMC password field empty",
			&serverserviceapi.Server{Attributes: []serverserviceapi.Attributes{{Namespace: bmcAttributeNamespace}}},
			&serverserviceapi.ServerCredential{Username: "foo", Password: ""},
			true,
			"BMC password field empty",
		},
		{
			"BMC username field empty",
			&serverserviceapi.Server{Attributes: []serverserviceapi.Attributes{{Namespace: bmcAttributeNamespace}}},
			&serverserviceapi.ServerCredential{Username: "", Password: "123"},
			true,
			"BMC username field empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRequiredAttributes(tc.server, tc.secret, tc.expectCredentials)
			if tc.expectedErr != "" {
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func Test_toAsset(t *testing.T) {
	cases := []struct {
		name          string
		server        *serverserviceapi.Server
		secret        *serverserviceapi.ServerCredential
		expectedAsset *model.Asset
		expectedErr   string
	}{
		{
			"Expected attributes empty raises error",
			&serverserviceapi.Server{
				Attributes: []serverserviceapi.Attributes{
					{
						Namespace: "invalid",
					},
				},
			},
			&serverserviceapi.ServerCredential{Username: "foo", Password: "bar"},
			nil,
			"expected server attributes with BMC address, got none",
		},
		{
			"Attributes missing BMC IP Address raises error",
			&serverserviceapi.Server{
				Attributes: []serverserviceapi.Attributes{
					{
						Namespace: bmcAttributeNamespace,
						Data:      []byte(`{"namespace":"foo"}`),
					},
				},
			},
			&serverserviceapi.ServerCredential{Username: "user", Password: "hunter2"},
			nil,
			"expected BMC address attribute empty",
		},
		{
			"Valid server, secret objects returns *model.Asset object",
			&serverserviceapi.Server{
				Attributes: []serverserviceapi.Attributes{
					{
						Namespace: bmcAttributeNamespace,
						Data:      []byte(`{"address":"127.0.0.1"}`),
					},
				},
			},
			&serverserviceapi.ServerCredential{Username: "user", Password: "hunter2"},
			&model.Asset{
				ID:          "00000000-0000-0000-0000-000000000000",
				Vendor:      "unknown",
				Model:       "unknown",
				Serial:      "unknown",
				Facility:    "",
				BMCUsername: "user",
				BMCPassword: "hunter2",
				BMCAddress:  net.ParseIP("127.0.0.1"),
				Metadata:    map[string]string{},
			},
			"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			asset, err := toAsset(tc.server, tc.secret, true)
			if tc.expectedErr != "" {
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedAsset, asset)
		})
	}
}

func Test_vendorDataUpdate(t *testing.T) {
	type args struct {
		new     map[string]string
		current map[string]string
	}

	// nolint:govet // test code is test code - disable struct fieldalignment error
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			"current is nil",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
				current: nil,
			},
			map[string]string{
				serverSerialAttributeKey: "01234",
				serverVendorAttributeKey: "foo",
				serverModelAttributeKey:  "bar",
			},
		},
		{
			"current and new data is equal",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
				current: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
			},
			nil,
		},
		{
			"current empty attribute is updated",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
				current: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "",
					serverModelAttributeKey:  "bar",
				},
			},
			map[string]string{
				serverSerialAttributeKey: "01234",
				serverVendorAttributeKey: "foo",
				serverModelAttributeKey:  "bar",
			},
		},
		{
			"current unknown and empty attributes are updated",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
				current: map[string]string{
					serverSerialAttributeKey: "unknown",
					serverVendorAttributeKey: "",
					serverModelAttributeKey:  "bar",
				},
			},
			map[string]string{
				serverSerialAttributeKey: "01234",
				serverVendorAttributeKey: "foo",
				serverModelAttributeKey:  "bar",
			},
		},
		{
			"current attributes are not updated",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234LLL",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
				current: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
			},
			nil,
		},
		{
			"current attributes are not updated - with unknown value",
			args{
				new: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "unknown",
					serverModelAttributeKey:  "bar",
				},
				current: map[string]string{
					serverSerialAttributeKey: "01234",
					serverVendorAttributeKey: "foo",
					serverModelAttributeKey:  "bar",
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vendorDataUpdate(tt.args.new, tt.args.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func assertComponentAttributes(t *testing.T, obj *serverserviceapi.ServerComponent, expectedVersion string) {
	t.Helper()

	assert.NotNil(t, obj)
	assert.NotNil(t, obj.ServerUUID)
	assert.NotNil(t, obj.UUID)
	assert.NotNil(t, obj.ComponentTypeSlug)
	assert.NotEmpty(t, obj.VersionedAttributes[0].Data)
	assert.True(t, rawVersionAttributeFirmwareEquals(t, expectedVersion, obj.VersionedAttributes[0].Data))
}

// rawVersionAttributeKVEquals returns a bool value when the given key and value is equal
func rawVersionAttributeFirmwareEquals(t *testing.T, expectedVersion string, rawVA []byte) bool {
	t.Helper()

	va := &versionedAttributes{}

	err := json.Unmarshal(rawVA, va)
	if err != nil {
		t.Fatal(err)
	}

	return va.Firmware.Installed == expectedVersion
}

func Test_ServerServiceChangeList(t *testing.T) {
	components := fixtures.CopyServerServiceComponentSlice(fixtures.ServerServiceR6515Components_fc167440)

	// nolint:govet // struct alignment kept for readability
	testcases := []struct {
		name            string // test name
		current         []*serverserviceapi.ServerComponent
		expectedUpdate  int
		expectedAdd     int
		expectedRemove  int
		slug            string // the component slug
		vaUpdates       *versionedAttributes
		aUpdates        *attributes
		addComponent    bool // adds a new component into the new slice before comparison
		removeComponent bool // removes a component from the new slice
	}{
		{
			"no changes in component lists",
			componentPtrSlice(fixtures.CopyServerServiceComponentSlice(components)),
			0,
			0,
			0,
			"",
			nil,
			nil,
			false,
			false,
		},
		{
			"updated component part of update slice",
			componentPtrSlice(fixtures.CopyServerServiceComponentSlice(components)),
			1,
			0,
			0,
			common.SlugBIOS,
			&versionedAttributes{Firmware: &common.Firmware{Installed: "2.2.6"}},
			nil,
			false,
			false,
		},
		{
			"added component part of add slice",
			componentPtrSlice(fixtures.CopyServerServiceComponentSlice(components)),
			0,
			1,
			0,
			common.SlugNIC,
			&versionedAttributes{Firmware: &common.Firmware{Installed: "1.3.3"}},
			nil,
			true,
			false,
		},
		{
			"component removed from slice",
			componentPtrSlice(fixtures.CopyServerServiceComponentSlice(components)),
			0,
			0,
			1,
			"",
			nil,
			nil,
			false,
			true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			newObjs := componentPtrSlice(fixtures.CopyServerServiceComponentSlice(fixtures.ServerServiceR6515Components_fc167440))

			switch {
			case tc.expectedAdd > 0:
				newObjs = addcomponent(newObjs, t, tc.slug, tc.vaUpdates)
			case tc.expectedUpdate > 0:
				newObjs = updateComponentVA(newObjs, t, tc.slug, tc.vaUpdates)
			case tc.expectedRemove > 0:
				newObjs = newObjs[:len(newObjs)-1]
			default:
			}

			gotAdd, gotUpdate, gotRemove, err := serverServiceChangeList(context.TODO(), tc.current, newObjs)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.expectedAdd, len(gotAdd), "add list differs")
			assert.Equal(t, tc.expectedUpdate, len(gotUpdate), "update list differs")
			assert.Equal(t, tc.expectedRemove, len(gotRemove), "remove list differs")
		})
	}
}
