SHELL=/bin/bash

BLACK        := $(shell tput -Tansi setaf 0)
RED          := $(shell tput -Tansi setaf 1)
GREEN        := $(shell tput -Tansi setaf 2)
YELLOW       := $(shell tput -Tansi setaf 3)
LIGHTPURPLE  := $(shell tput -Tansi setaf 4)
PURPLE       := $(shell tput -Tansi setaf 5)
BLUE         := $(shell tput -Tansi setaf 6)
WHITE        := $(shell tput -Tansi setaf 7)

BG_BLACK        := $(shell tput -Tansi setab 0)
BG_RED          := $(shell tput -Tansi setab 1)
BG_GREEN        := $(shell tput -Tansi setab 2)
BG_YELLOW       := $(shell tput -Tansi setab 3)
BG_LIGHTPURPLE  := $(shell tput -Tansi setab 4)
BG_PURPLE       := $(shell tput -Tansi setab 5)
BG_BLUE         := $(shell tput -Tansi setab 6)
BG_WHITE        := $(shell tput -Tansi setab 7)

BOLD            := $(shell tput -Tansi bold)
DIM             := $(shell tput -Tansi dim)
UNDERLINE_START := $(shell tput -Tansi smul)
UNDERLINE_END   := $(shell tput -Tansi rmul)
INVERT          := $(shell tput -Tansi rev)
STANDOUT_START  := $(shell tput -Tansi smso)
STANDOUT_END    := $(shell tput -Tansi rmso)

RESET := $(shell tput -Tansi sgr0)

define NEWLINE

endef

define FONT_URLS
$(shell grep -irn "for use with" thirdparty/IconFontCppHeaders | grep http | awk '{print $$NF}' | sort -u | xargs)
endef

define FONT_URL_COUNT
$(shell echo ${FONT_URLS} | sed 's/ /\n/g' | wc -l)
endef


TARGET_COLOR := $(LIGHTPURPLE)

POUND = \#

.PHONY: default
default:banner no-target-warning update update-submodules help;
no-target-warning:
	@echo "${BOLD}${PURPLE}Target not set! Updating project instead ...${RESET}"

update:
	@echo "${BG_LIGHTPURPLE}Updating repository ...${RESET}"
	@git pull origin main --quiet
	@echo

update-submodules: ## Updates all submodules defined in .gitmodules
	@echo "${BG_LIGHTPURPLE}Init submodules if needed ...${RESET}"
	@git submodule update --init --quiet
	@echo
	@echo "${BG_LIGHTPURPLE}Pulling submodule updates if needed ...${RESET}"
	@git submodule foreach --quiet git pull origin main
	@git submodule foreach --quiet git checkout main
	@echo

pre-submodule-update:
	@echo "${BG_LIGHTPURPLE}Removing previously downloaded fonts ...${RESET}";
	find resources/fonticons/ -type f -iname '*ttf*' -exec rm -v {} \;
	find resources/fonticons/ -type f -iname '*.go' -exec rm -v {} \;
	@echo

help: ## Show this screen
	@echo "Run ${BOLD}${LIGHTPURPLE}make help${RESET}${RESET} for a list of commands."
	@echo
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "${TARGET_COLOR}%-30s${RESET} %s\n", $$1, $$2}'
	@echo


update-font-icons: ## Download fonts, generate headers and move files to src.
	@echo "${BG_LIGHTPURPLE}Copying GenerateIconFontCppHeaders.py as module locally ...${RESET}";
	@grep -irn "# Main" thirdparty/IconFontCppHeaders/GenerateIconFontCppHeaders.py | awk -F ":" '{print "cat thirdparty/IconFontCppHeaders/GenerateIconFontCppHeaders.py | head -n " $$1 " > scripts/icon_font_cpp_headers.py"}' | sh -v;
	@touch scripts/__init__.py
	@echo
	@echo "${BG_LIGHTPURPLE}Running wrapper script ...${RESET}";
	@python3 scripts/GenerateIconFontCppHeaders_wrapper.py
	@echo
	@echo "${BG_LIGHTPURPLE}Moving files ...${RESET}";
	@mv -v Icons*.go resources/fonticons/
	@echo
	@echo "${BG_LIGHTPURPLE}Copying go module ...${RESET}";
	@cp -v thirdparty/IconFontCppHeaders/font.go resources/fonticons/
	@echo
	@echo "${BG_LIGHTPURPLE}Cleaning up generated scripts ...${RESET}";
	@rm -v scripts/icon_font_cpp_headers.py
	@rm -v scripts/__init__.py

build-init:
	@echo "${BG_LIGHTPURPLE}Build init ...${RESET}";
	@mkdir -p build

build-clean: ## Clean up build directory contents
	@echo "${BG_LIGHTPURPLE}Cleaning up previous builds ...${RESET}";
	find build/ -type f -exec rm -v {} \;

build: build-init build-clean ## Build the project
	@echo "${BG_LIGHTPURPLE}Building from src ...${RESET}";
	@go build -o build/bbe src/main.go

build-run: build ## Build and run the project
	@echo "${BG_LIGHTPURPLE}Running ...${RESET}";
	@build/bbe
