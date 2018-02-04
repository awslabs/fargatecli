package cmd

import "errors"

var consoleOutput = ConsoleOutput{
	Color:   false,
	Emoji:   false,
	Verbose: true,
	Test:    true,
}

func ExampleConsoleOutput_Debug() {
	consoleOutput.Debug("PC LOAD LETTER")
	// Output: [d] PC LOAD LETTER
}

func ExampleConsoleOutput_Info() {
	consoleOutput.Info("Welcome! Everything is %s.", "fine")
	// Output: [i] Welcome! Everything is fine.
}

func ExampleConsoleOutput_Fatal() {
	err := errors.New("OXY2_TANK_EXPLOSION")
	consoleOutput.Fatal(err, "Houston, we've had a problem.")
	// Output:
	// [!] Houston, we've had a problem.
	//     OXY2_TANK_EXPLOSION
}

func ExampleConsoleOutput_Fatals() {
	errs := []error{
		errors.New("OXY2_TANK_EXPLOSION"),
		errors.New("PRIM_FUEL_CELL_FAILURE"),
		errors.New("SEC_FUEL_CELL_FAILURE"),
	}
	consoleOutput.Fatals(errs, "Houston, we've had a problem.")
	// Output:
	// [!] Houston, we've had a problem.
	//     OXY2_TANK_EXPLOSION
	//     PRIM_FUEL_CELL_FAILURE
	//     SEC_FUEL_CELL_FAILURE
}

func ExampleConsoleOutput_KeyValue() {
	population := 468730

	consoleOutput.KeyValue("Name", "Staten Island", 0)
	consoleOutput.KeyValue("County", "Richmond", 1)
	consoleOutput.KeyValue("Population", "%d", 1, population)
	// Output:
	// Name: Staten Island
	//     County: Richmond
	//     Population: 468730
}

func ExampleConsoleOutput_Say() {
	username := "Werner Brandes"
	consoleOutput.Say("Hi, my name is %s. My voice is my passport. Verify Me.", 0, username)
	// Output: Hi, my name is Werner Brandes. My voice is my passport. Verify Me.
}

func ExampleConsoleOutput_Warn() {
	consoleOutput.Warn("Keep it secret, keep it safe.")
	// Output: [!] Keep it secret, keep it safe.
}

func ExampleConsoleOutput_Table() {
	rows := [][]string{
		{"NAME", "ALLEGIANCE"},
		{"Butterbumps", "House Tyrell"},
		{"Jinglebell", "House Frey"},
		{"Moon Boy", "House Baratheon"},
	}

	consoleOutput.Table("Fools of Westeros", rows)
	// Output:
	// Fools of Westeros
	//
	// NAME		ALLEGIANCE
	// Butterbumps	House Tyrell
	// Jinglebell	House Frey
	// Moon Boy	House Baratheon
}
