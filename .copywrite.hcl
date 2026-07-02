schema_version = 1

project {
  license          = "BUSL-1.1"
  copyright_holder = "Dit"
  copyright_year   = 2026

  # Generated and non-source files that must not carry license headers.
  header_ignore = [
    # protoc output — regeneration would strip a manually added header
    "internal/proto/*.pb.go",

    # build artifacts and coverage output
    "build/**",
    ".health/**",
    "**/*.out",
  ]
}
