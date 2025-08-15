package cmd

import (
	"fmt"
	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/spf13/cobra"
)

var testPrint = &cobra.Command{
	Use:   "test_print",
	Short: "Print a large DataFrame to test formatting",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataframe := df.Dataframe{
			ColumnOrder: make([]string, 0),
			Columns:     make(map[string]df.ColumnInterface),
		}

		// Create sample data
		ids := []string{}
		names := []string{}
		genders := []string{}

		for i := 1; i <= 25; i++ {
			ids = append(ids, fmt.Sprintf("%03d", i))

			// Add a very long name every 5th row
			if i%5 == 0 {
				names = append(names, fmt.Sprintf("Alan aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa%d", i))
			} else {
				names = append(names, fmt.Sprintf("Name%d", i))
			}

			// Alternate genders
			if i%2 == 0 {
				genders = append(genders, "M")
			} else {
				genders = append(genders, "F")
			}
		}

		// Add columns
		dataframe.NewColumn("Id", df.NewColumn("Id", ids))
		dataframe.NewColumn("Name", df.NewColumn("Name", names))
		dataframe.NewColumn("Gender", df.NewColumn("Gender", genders))

		// Print the DataFrame
		fmt.Print(dataframe)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testPrint)
}
