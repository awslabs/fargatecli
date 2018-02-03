package cmd

import "testing"

var validateCpuAndMemoryTests = []struct {
	CpuUnits  string
	Mebibytes string
	Out       error
}{
	// 0.25 vCpu
	{"256", "512", nil},
	{"256", "1024", nil},
	{"256", "2048", nil},
	{"256", "0", InvalidCpuAndMemoryCombination},
	{"256", "768", InvalidCpuAndMemoryCombination},
	{"256", "2151", InvalidCpuAndMemoryCombination},
	{"256", "3072", InvalidCpuAndMemoryCombination},

	// 0.5 vCpu
	{"512", "1024", nil},
	{"512", "2048", nil},
	{"512", "3072", nil},
	{"512", "4096", nil},
	{"512", "512", InvalidCpuAndMemoryCombination},

	// 1 vCpu
	{"1024", "2048", nil},
	{"1024", "5120", nil},
	{"1024", "8192", nil},
	{"1024", "1024", InvalidCpuAndMemoryCombination},
	{"1024", "9216", InvalidCpuAndMemoryCombination},

	// 2 vCpu
	{"2048", "4096", nil},
	{"2048", "10240", nil},
	{"2048", "16384", nil},
	{"2048", "3072", InvalidCpuAndMemoryCombination},
	{"2048", "17408", InvalidCpuAndMemoryCombination},

	// 4 vCpu
	{"4096", "8192", nil},
	{"4096", "15360", nil},
	{"4096", "30720", nil},
	{"4096", "1024", InvalidCpuAndMemoryCombination},
	{"4096", "31744", InvalidCpuAndMemoryCombination},
}

func TestValidateCpuAndMemoryWithValidParameters(t *testing.T) {
	err := validateCpuAndMemory("256", "512")

	if err != nil {
		t.Errorf("Validation failed, got %s, want nil", err)
	}
}

func TestValidateCpuAndMemoryWithInvalidParameters(t *testing.T) {
	err := validateCpuAndMemory("5", "23849")

	if err == nil {
		t.Errorf("Validation failed, got nil, want %s", InvalidCpuAndMemoryCombination)
	}
}

func TestValidateCpuAndMemory(t *testing.T) {
	for _, test := range validateCpuAndMemoryTests {
		s := validateCpuAndMemory(test.CpuUnits, test.Mebibytes)

		if s != test.Out {
			t.Errorf("validateCpuAndMemory(%s, %s) => %#v, want %s", test.CpuUnits, test.Mebibytes, s, test.Out)
		}
	}
}
