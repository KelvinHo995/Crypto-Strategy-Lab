package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFilesPreservesProcessEnvironment(t *testing.T) {
	const fileOnlyKey = "CRYPTO_STRATEGY_LAB_DOTENV_TEST"
	oldValue, existed := os.LookupEnv(fileOnlyKey)
	if err := os.Unsetenv(fileOnlyKey); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(fileOnlyKey, oldValue)
		} else {
			_ = os.Unsetenv(fileOnlyKey)
		}
	})

	t.Setenv("JWT_SECRET", "process-environment-secret")
	path := filepath.Join(t.TempDir(), ".env")
	contents := fileOnlyKey + "=loaded-from-file\nJWT_SECRET=file-secret\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	loadEnvFiles(path)

	if got := os.Getenv(fileOnlyKey); got != "loaded-from-file" {
		t.Fatalf("file value = %q, want loaded-from-file", got)
	}
	if got := os.Getenv("JWT_SECRET"); got != "process-environment-secret" {
		t.Fatalf("JWT_SECRET = %q, process environment should take precedence", got)
	}
}
