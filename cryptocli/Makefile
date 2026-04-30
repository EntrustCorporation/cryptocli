#
# Copyright (c) 2023 Entrust, Inc. All Rights Reserved.
#
THISDIR = $(shell pwd)

GOCMD = /usr/local/go/bin/go

CRYPTOCLI_SRCDIR = $(THISDIR)
PARENTDIR = $(THISDIR)/..

BUILDDIR = cryptocli-build
WORKSPACE = $(PARENTDIR)/$(BUILDDIR)
WORKSPACE_SRCDIR = $(WORKSPACE)/src
WORKSPACE_BINDIR = $(WORKSPACE)/bin

CRYPTOCLI_MAINGO = main.go

CRYPTOCLI_LINUX = cryptocli

all:
	@/usr/bin/echo "Clearing old Workspace if any at $(WORKSPACE).."
	@/usr/bin/rm -rf $(WORKSPACE)
	@/usr/bin/echo "Creating new cryptocli Workspace at $(WORKSPACE).."
	@/usr/bin/mkdir $(WORKSPACE)
	@/usr/bin/mkdir $(WORKSPACE_SRCDIR)
	@/usr/bin/cp -r $(CRYPTOCLI_SRCDIR)/. $(WORKSPACE_SRCDIR)
	@/usr/bin/cp -r $(PARENTDIR)/getpasswd/ $(WORKSPACE_SRCDIR)/
	@/usr/bin/cp -r $(PARENTDIR)/cmd/. $(WORKSPACE_SRCDIR)/cmd/
	@cd $(WORKSPACE_SRCDIR) && $(GOCMD) mod init cli
	@cd $(WORKSPACE_SRCDIR) && $(GOCMD) mod tidy
	@/usr/bin/echo "Compiling cryptocli for Linux..."
	@cd $(WORKSPACE_SRCDIR) && (env GOOS=linux GOARCH=amd64 $(GOCMD) build -o $(WORKSPACE_BINDIR)/$(CRYPTOCLI_LINUX) $(CRYPTOCLI_MAINGO))
	@/usr/bin/echo "Please find Linux cryptocli binary, $(CRYPTOCLI_LINUX) at $(WORKSPACE_BINDIR)"

install:

clean:
	@/usr/bin/echo "Clearing old Workspace if any at $(WORKSPACE).."
	@/usr/bin/rm -rf $(WORKSPACE)
	@/usr/bin/echo "Clean up complete..."
