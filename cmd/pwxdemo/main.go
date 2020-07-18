package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thockin/go-build-template/pkg/pwx"
)

var (
	rootCmd = &cobra.Command{
		Use:   "pwxdemo",
		Short: "A tool to test pwx apis",
		Long:  `Use this to do magic`,
	}

	volumeID     string
	takeSnapshot = &cobra.Command{
		Use:   "snapshot",
		Short: "Take a snapshot of a volume",
		Long:  `Take a snapshot`,
		Run: func(cmd *cobra.Command, args []string) {
			if snapID, err := pwx.CreateSnapshot(volumeID); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("Created Snapshot - [%s]\n", snapID)
			}
		},
	}

	name, storageclass, snapshotid, namespace string

	restore = &cobra.Command{
		Use:   "restore",
		Short: "Restore a volume",
		Long:  `Restore a volume`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := pwx.CreatePVCFromSnapshot(name, storageclass, snapshotid, namespace); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Restore succeeded")
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(takeSnapshot)
	takeSnapshot.Flags().StringVarP(&volumeID, "volumeid", "v", "", "The ID of a volume")

	rootCmd.AddCommand(restore)
	restore.Flags().StringVarP(&name, "name", "n", "", "The pvc/pv name gen")
	restore.Flags().StringVarP(&storageclass, "storageclass", "s", "", "The storageclass")
	restore.Flags().StringVarP(&snapshotid, "snapshotid", "i", "", "The snapshot id")
	restore.Flags().StringVarP(&namespace, "namespace", "m", "", "The namespace")
}

func main() {
	Execute()
}

// Execute the thing
func Execute() error {
	return rootCmd.Execute()
}
