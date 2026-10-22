package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceHandlersAcrossControllers(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "multi.controller.ts")
	source := `@Controller('platform/members')
export class MembersController {
  @Get()
  list() { return 1; }
}
@Controller('platform/invitations')
export class InvitationsController {
  @Post('accept')
  async accept() { return 2; }
}`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	handlers, err := sourceHandlers(root, "multi.controller.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := handlers["platform/members:GET:"]; len(got) != 1 || got[0] != "list" {
		t.Fatalf("first controller handlers: %+v", handlers)
	}
	if got := handlers["platform/invitations:POST:accept"]; len(got) != 1 || got[0] != "accept" {
		t.Fatalf("second controller handlers: %+v", handlers)
	}
}

func TestAuditRejectsMissingAndDuplicateMappings(t *testing.T) {
	root, source := t.TempDir(), t.TempDir()
	sourceText := "@Controller('api')\nexport class TestController {\n @Get('one')\n one() {}\n @Post('two')\n two() {}\n}"
	if err := os.WriteFile(filepath.Join(source, "source.ts"), []byte(sourceText), 0600); err != nil {
		t.Fatal(err)
	}
	inventory := []route{
		{ID: "JZ-ROUTE-001", SourceFile: "source.ts", ControllerPath: "api", MethodPath: "one", SourceHandler: "one", Method: "GET", Path: "api/one", Implementation: "UNAUDITED"},
		{ID: "JZ-ROUTE-002", SourceFile: "source.ts", ControllerPath: "api", MethodPath: "two", SourceHandler: "two", Method: "POST", Path: "api/two", Implementation: "UNAUDITED"},
	}
	if err := validate(root, source, inventory, inventory[:1]); err == nil {
		t.Fatal("missing route mapping accepted")
	}
	if err := validate(root, source, inventory, []route{inventory[0], inventory[0]}); err == nil {
		t.Fatal("duplicate route ID accepted")
	}
	broken := append([]route(nil), inventory...)
	broken[0].Implementation = "PARTIAL"
	broken[0].DestinationEntry = "jazebe_fake"
	broken[0].DestinationFile = "missing.go"
	if err := validate(root, source, inventory, broken); err == nil {
		t.Fatal("missing destination evidence accepted")
	}
	broken = append([]route(nil), inventory...)
	broken[0].Tests = []string{"missing_test.go (NOT_RUN)"}
	if err := validate(root, source, inventory, broken); err == nil {
		t.Fatal("stale test evidence accepted")
	}
	if err := validate(root, source, inventory, inventory); err != nil {
		t.Fatalf("complete unreviewed inventory rejected: %v", err)
	}
}
