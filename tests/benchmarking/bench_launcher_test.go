package benchmarking

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"testing"
)

var launcher *lbconfig.AppLauncher

func init() {
	launcher = lbconfig.NewAppLauncher()
}

// BenchmarkLauncherParsing : benchmarking of the application arguments
func BenchmarkLauncherParsing(b *testing.B) {
	args1 := []string{""}
	args2 := []string{"-v"}
	args3 := []string{"--rotatecfg.enabled", "-d DEBUG", "-v"}
	args4 := []string{"--rotatecfg.enabled", "-d DEBUG", "-v", "-d FATAL", "-l log/app.log"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.Run("test-group", func(b *testing.B) {
			b.Run("0", callArgsParsing(args1))
			b.Run("1", callArgsParsing(args2))
			b.Run("3", callArgsParsing(args3))
			b.Run("5", callArgsParsing(args4))
		})
	}
}

func BenchmarkLauncherRun(b *testing.B) {
	// Find the default argument values before running the launcher
	err := launcher.ParseApplicationArguments([]string{})
	if err != nil {
		b.Fatal(err.Error())
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = launcher.Run()
	}
}

/* Helpers */
func callArgsParsing(args []string) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		err := launcher.ParseApplicationArguments(args)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}