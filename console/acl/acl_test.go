// Copyright 2024 Heroic Labs.
// All rights reserved.
//
// NOTICE: All information contained herein is, and remains the property of Heroic
// Labs. and its suppliers, if any. The intellectual and technical concepts
// contained herein are proprietary to Heroic Labs. and its suppliers and may be
// covered by U.S. and Foreign Patents, patents in process, and are protected by
// trade secret or copyright law. Dissemination of this information or reproduction
// of this material is strictly forbidden unless prior written permission is
// obtained from Heroic Labs.

package acl

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/heroiclabs/nakama/v3/console"
	"github.com/stretchr/testify/assert"
)

func Test_Permission(t *testing.T) {
	p := NewPermission(console.AclResources_ACCOUNT, PermissionRead).
		Compose(NewPermission(console.AclResources_ACCOUNT, PermissionWrite)).
		Compose(NewPermission(console.AclResources_ACCOUNT_WALLET, PermissionRead)).
		Compose(NewPermission(console.AclResources_ACCOUNT_EXPORT, PermissionDelete))

	t.Logf("Permission bits: %s", p.bitmapString())

	acl := p.ACL()

	assert.True(t, p.HasAccess(NewPermission(console.AclResources_ACCOUNT, PermissionRead)))
	assert.True(t, p.HasAccess(NewPermission(console.AclResources_ACCOUNT, PermissionRead)))
	assert.True(t, p.HasAccess(NewPermission(console.AclResources_ACCOUNT, PermissionWrite)))
	assert.True(t, p.HasAccess(NewPermission(console.AclResources_ACCOUNT_EXPORT, PermissionDelete)))
	assert.False(t, p.HasAccess(NewPermission(console.AclResources_ACCOUNT, PermissionDelete)))
	assert.True(t, acl[console.AclResources_ACCOUNT.String()].Read)
	assert.True(t, acl[console.AclResources_ACCOUNT.String()].Write)
	assert.False(t, acl[console.AclResources_ACCOUNT.String()].Delete)
	assert.True(t, acl[console.AclResources_ACCOUNT_WALLET.String()].Read)
	assert.False(t, acl[console.AclResources_ACCOUNT_WALLET.String()].Write)
	assert.False(t, acl[console.AclResources_ACCOUNT_WALLET.String()].Delete)
	assert.False(t, acl[console.AclResources_ACCOUNT_EXPORT.String()].Read)
	assert.False(t, acl[console.AclResources_ACCOUNT_EXPORT.String()].Write)
	assert.True(t, acl[console.AclResources_ACCOUNT_EXPORT.String()].Delete)
}

func TestCheckACLHttpEconomy(t *testing.T) {
	read := NewPermission(console.AclResources_HIRO_ECONOMY, PermissionRead)
	write := NewPermission(console.AclResources_HIRO_ECONOMY, PermissionWrite)
	if !CheckACLHttp(http.MethodGet, "/v2/console/economy/products", read) {
		t.Fatal("read permission must list products")
	}
	if CheckACLHttp(http.MethodPost, "/v2/console/economy/products", read) {
		t.Fatal("read permission must not create products")
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		if !CheckACLHttp(method, "/v2/console/economy/products/{id}", write) {
			t.Fatalf("write permission must allow %s", method)
		}
	}
	if CheckACLHttp(http.MethodGet, "/v2/console/economy/products", None()) {
		t.Fatal("no permission must be refused")
	}
	if !CheckACLHttp(http.MethodDelete, "/v2/console/economy/gifts/{id}", Admin()) {
		t.Fatal("admin must pass")
	}
	if CheckACLHttp("", "/v2/console/economy/gifts/{id}", write) {
		t.Fatal("handlers without a single method need admin")
	}
}

func TestCheckACLHttpPublication(t *testing.T) {
	read := NewPermission(console.AclResources_NOTIFICATION, PermissionRead)
	write := NewPermission(console.AclResources_NOTIFICATION, PermissionWrite)
	if !CheckACLHttp(http.MethodGet, "/v2/console/publication/news", read) {
		t.Fatal("read permission must list news")
	}
	if CheckACLHttp(http.MethodPost, "/v2/console/publication/notifications", read) {
		t.Fatal("read permission must not send notifications")
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		if !CheckACLHttp(method, "/v2/console/publication/providers/{id}", write) {
			t.Fatalf("write permission must allow %s", method)
		}
	}
	if CheckACLHttp(http.MethodGet, "/v2/console/publication/faqs", NewPermission(console.AclResources_HIRO_ECONOMY, PermissionRead)) {
		t.Fatal("economy permission must not read publication")
	}
	if CheckACLHttp(http.MethodGet, "/v2/console/publication/faqs", None()) {
		t.Fatal("no permission must be refused")
	}
	if !CheckACLHttp(http.MethodPost, "/v2/console/publication/providers/{id}/test", Admin()) {
		t.Fatal("admin must pass")
	}
	if CheckACLHttp("", "/v2/console/publication/faqs/{id}", write) {
		t.Fatal("handlers without a single method need admin")
	}
}

func TestCheckACLHttpBase(t *testing.T) {
	settingsRead := NewPermission(console.AclResources_SETTINGS, PermissionRead)
	settingsWrite := NewPermission(console.AclResources_SETTINGS, PermissionWrite)
	economyRead := NewPermission(console.AclResources_HIRO_ECONOMY, PermissionRead)
	economyWrite := NewPermission(console.AclResources_HIRO_ECONOMY, PermissionWrite)
	for _, path := range []string{"/v2/console/base/languages", "/v2/console/base/languages/{id}", "/v2/console/base/languages/coverage"} {
		if !CheckACLHttp(http.MethodGet, path, settingsRead) {
			t.Fatalf("settings read must read %s", path)
		}
		if CheckACLHttp(http.MethodGet, path, economyRead) {
			t.Fatalf("economy permission must not read %s", path)
		}
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		if !CheckACLHttp(method, "/v2/console/base/languages/{id}", settingsWrite) {
			t.Fatalf("settings write must allow %s on languages", method)
		}
		if CheckACLHttp(method, "/v2/console/base/languages/{id}", settingsRead) {
			t.Fatalf("settings read must not allow %s on languages", method)
		}
		if CheckACLHttp(method, "/v2/console/base/languages/{id}", economyWrite) {
			t.Fatalf("economy write must not allow %s on languages", method)
		}
	}
	for _, path := range []string{"/v2/console/base/categories", "/v2/console/base/categories/{id}", "/v2/console/base/tags/key-types", "/v2/console/base/tags/{id}"} {
		if !CheckACLHttp(http.MethodGet, path, economyRead) {
			t.Fatalf("economy read must read %s", path)
		}
		if CheckACLHttp(http.MethodGet, path, settingsRead) {
			t.Fatalf("settings permission must not read %s", path)
		}
		if CheckACLHttp(http.MethodPost, path, economyRead) {
			t.Fatalf("economy read must not write %s", path)
		}
		if !CheckACLHttp(http.MethodPut, path, economyWrite) {
			t.Fatalf("economy write must write %s", path)
		}
	}
	if CheckACLHttp(http.MethodGet, "/v2/console/base/categories", None()) || CheckACLHttp(http.MethodGet, "/v2/console/base/other", economyRead) {
		t.Fatal("unknown base paths and no permission must be refused")
	}
	if !CheckACLHttp(http.MethodDelete, "/v2/console/base/languages/{id}", Admin()) {
		t.Fatal("admin must pass")
	}
	if CheckACLHttp("", "/v2/console/base/tags/{id}", economyWrite) {
		t.Fatal("handlers without a single method need admin")
	}
}

func Test_AclUpgrades(t *testing.T) {
	limited := NewPermission(console.AclResources_ACCOUNT, PermissionRead).
		Compose(NewPermission(console.AclResources_USER, PermissionWrite)).
		Compose(NewPermission(console.AclResources_HIRO_STATS, PermissionDelete))

	lastResource := ConsoleResource(len(console.AclResources_value) - 1)
	allButLastResourceDelete := None()
	for i := range len(console.AclResources_value) {
		for _, level := range []PermissionLevel{PermissionRead, PermissionWrite, PermissionDelete} {
			if ConsoleResource(i) == lastResource && level == PermissionDelete {
				continue
			}
			allButLastResourceDelete = allButLastResourceDelete.Compose(NewPermission(ConsoleResource(i), level))
		}
	}

	// A newer version's token: the same user plus permissions on resources this version does not
	// know, which sit in this version's trailing padding bits and in an extra byte.
	newerLimited := append(bytes.Clone(limited.bitmap), 0xE0)
	// Padding is the unused low bits of the last byte; 1<<n - 1 sets exactly those n bits.
	paddingBits := byteCount*8 - len(console.AclResources_value)*3
	newerLimited[byteCount-1] |= byte(1<<paddingBits - 1)

	tests := []struct {
		name string
		in   []byte
		want Permission
	}{
		{name: "current length limited user is kept", in: limited.bitmap, want: limited},
		{name: "current length admin is kept", in: Admin().bitmap, want: Admin()},
		{name: "current length missing last resource delete is not promoted to admin", in: allButLastResourceDelete.bitmap, want: allButLastResourceDelete},
		{name: "older shorter limited user is zero-extended", in: limited.bitmap[:byteCount-1], want: limited},
		{name: "older shorter admin stays admin", in: Admin().bitmap[:byteCount-1], want: Admin()},
		{name: "much older shorter admin stays admin", in: Admin().bitmap[:byteCount-2], want: Admin()},
		{name: "empty bitmap has no access", in: nil, want: None()},
		{name: "newer longer limited user is truncated", in: newerLimited, want: limited},
		{name: "newer longer admin stays admin", in: append(Admin().bitmap, 0xFF), want: Admin()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.String(), NewFromBytes(tt.in).String())
		})
	}
}

func Test_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		in   Permission
		want bool
	}{
		{name: "admin is admin", in: Admin(), want: true},
		{name: "none is not admin", in: None(), want: false},
		{name: "zero value is not admin", in: Permission{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.IsAdmin())
		})
	}
}

func TestCheckACLHttpBaseBrandSetup(t *testing.T) {
	settingsRead := NewPermission(console.AclResources_SETTINGS, PermissionRead)
	settingsWrite := NewPermission(console.AclResources_SETTINGS, PermissionWrite)
	economyWrite := NewPermission(console.AclResources_HIRO_ECONOMY, PermissionWrite)
	for _, path := range []string{"/v2/console/base/brand", "/v2/console/base/setup"} {
		if !CheckACLHttp(http.MethodGet, path, settingsRead) || CheckACLHttp(http.MethodGet, path, economyWrite) {
			t.Fatalf("GET %s must need settings read", path)
		}
		if !CheckACLHttp(http.MethodPut, path, settingsWrite) || CheckACLHttp(http.MethodPut, path, settingsRead) {
			t.Fatalf("PUT %s must need settings write", path)
		}
	}
	if !CheckACLHttp(http.MethodPost, "/v2/console/base/setup/claim", settingsWrite) || CheckACLHttp(http.MethodPost, "/v2/console/base/setup/claim", settingsRead) {
		t.Fatal("claim must need settings write")
	}
}
