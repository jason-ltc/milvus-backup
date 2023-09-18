package cmd

import (
	"context"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/zilliztech/milvus-backup/core"
	"github.com/zilliztech/milvus-backup/core/paramtable"
	"github.com/zilliztech/milvus-backup/core/proto/backuppb"
	"github.com/zilliztech/milvus-backup/core/utils"
	"github.com/zilliztech/milvus-backup/internal/log"

	"go.uber.org/zap"
)

var (
	restoreBackupName          string
	restoreCollectionNames     string
	renameSuffix               string
	renameCollectionNames      string
	restoreDatabases           string
	restoreDatabaseCollections string
)

var restoreBackupCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore subcommand restore a backup.",

	Run: func(cmd *cobra.Command, args []string) {
		var params paramtable.BackupParams
		params.GlobalInitWithYaml(config)
		params.Init()

		context := context.Background()
		backupContext := core.CreateBackupContext(context, params)
		log.Info("restore cmd input args", zap.Strings("args", args))
		var collectionNameArr []string
		if restoreCollectionNames == "" {
			collectionNameArr = []string{}
		} else {
			collectionNameArr = strings.Split(restoreCollectionNames, ",")
		}

		renameMap := make(map[string]string, 0)
		if renameCollectionNames != "" {
			fmt.Println("rename: " + renameCollectionNames)
			renameArr := strings.Split(renameCollectionNames, ",")
			for _, rename := range renameArr {
				if strings.Contains(rename, ":") {
					splits := strings.Split(rename, ":")
					renameMap[splits[0]] = splits[1]
				} else {
					fmt.Println("illegal rename parameter")
					return
				}
			}
		}

		if restoreDatabaseCollections == "" && restoreDatabases != "" {
			dbCollectionDict := make(map[string][]string)
			splits := strings.Split(restoreDatabases, ",")
			for _, db := range splits {
				dbCollectionDict[db] = []string{}
			}
			completeDbCollections, err := jsoniter.MarshalToString(dbCollectionDict)
			restoreDatabaseCollections = completeDbCollections
			if err != nil {
				fmt.Println("illegal databases input")
				return
			}
		}
		resp := backupContext.RestoreBackup(context, &backuppb.RestoreBackupRequest{
			BackupName:        restoreBackupName,
			CollectionNames:   collectionNameArr,
			CollectionSuffix:  renameSuffix,
			CollectionRenames: renameMap,
			DbCollections:     utils.WrapDBCollections(restoreDatabaseCollections),
		})

		fmt.Println(resp.GetCode(), "\n", resp.GetMsg())
	},
}

func init() {
	restoreBackupCmd.Flags().StringVarP(&restoreBackupName, "name", "n", "", "backup name to restore")
	restoreBackupCmd.Flags().StringVarP(&restoreCollectionNames, "collections", "c", "", "collectionNames to restore")
	restoreBackupCmd.Flags().StringVarP(&renameSuffix, "suffix", "s", "", "add a suffix to collection name to restore")
	restoreBackupCmd.Flags().StringVarP(&renameCollectionNames, "rename", "r", "", "rename collections to new names, format: db1.collection1:db2.collection1_new,db1.collection2:db2.collection2_new")
	restoreBackupCmd.Flags().StringVarP(&restoreDatabases, "databases", "d", "", "databases to restore, if not set, restore all databases")
	restoreBackupCmd.Flags().StringVarP(&restoreDatabaseCollections, "database_collections", "a", "", "databases and collections to restore, json format: {\"db1\":[\"c1\", \"c2\"],\"db2\":[]}")

	rootCmd.AddCommand(restoreBackupCmd)
}
