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
