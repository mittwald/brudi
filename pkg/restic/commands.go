package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mittwald/brudi/pkg/cli"
)

const (
	binary             = "restic"
	fileType           = "file"
	messageType        = "message_type"
	snapshotID         = "snapshot_id"
	parentID           = "parent"
	messageTypeSummary = "summary"
)

var (
	ErrRepoAlreadyInitialized = fmt.Errorf("repo already initialized")

	cmdTimeout = 6 * time.Hour
)

// InitBackup executes "restic init"
func initBackup(ctx context.Context, globalOpts *GlobalOptions) ([]byte, error) {
	cmd := newCommand("init", cli.StructToCLI(globalOpts)...)

	out, err := cli.RunWithTimeout(ctx, cmd, cmdTimeout)
	if err != nil {
		// s3 init-check
		if strings.Contains(string(out), "config already initialized") {
			return out, ErrRepoAlreadyInitialized
		}
		// file init-check
		if strings.Contains(string(out), "file already exists") {
			return out, ErrRepoAlreadyInitialized
		}
		return out, err
	}
	return out, err
}

// parseSnapshotOut retrieves snapshot-id and, if available, parent-id from json logs
func parseSnapshotOut(jsonLog []byte) (BackupResult, error) {
	var result BackupResult

	var responseList []map[string]*interface{}
	jErr := json.Unmarshal(jsonLog, &responseList)
	if jErr != nil {
		return BackupResult{}, jErr
	}

	var parentSnapshotID string
	var curSnapshotID string
	for idx := range responseList {
		v := responseList[idx]
		if v[messageType] != nil && *v[messageType] == messageTypeSummary {
			if v[snapshotID] != nil {
				curSnapshotID = (*v[snapshotID]).(string)
			}
			if v[parentID] != nil {
				parentSnapshotID = (*v[parentID]).(string)
			}
		}
	}
	if parentSnapshotID != "" {
		result.ParentSnapshotID = parentSnapshotID
	}
	if curSnapshotID == "" {
		return BackupResult{}, fmt.Errorf("failed to parse snapshotID")
	}
	result.SnapshotID = curSnapshotID
	return result, nil
}

// CreateBackup executes "restic backup" and returns the parent snapshot id (if available) and the snapshot id
func CreateBackup(ctx context.Context, globalOpts *GlobalOptions, backupOpts *BackupOptions, unlock bool) (BackupResult, []byte, error) {
	var out []byte
	var err error

	if unlock {
		unlockOpts := UnlockOptions{
			Flags: &UnlockFlags{
				RemoveAll: false,
			},
		}
		out, err = Unlock(ctx, globalOpts, &unlockOpts)
		if err != nil {
			return BackupResult{}, out, err
		}
	}

	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(backupOpts)...)

	cmd := newCommand("backup", args...)

	out, err = cli.RunWithTimeout(ctx, cmd, cmdTimeout)
	if err != nil {
		return BackupResult{}, out, err
	}

	// transform output from restic into list of json elements
	out = []byte(fmt.Sprint("[" +
		strings.ReplaceAll(strings.TrimRight(string(out), "\n"), "\n", ",") +
		"]"))

	var backupRes BackupResult
	backupRes, err = parseSnapshotOut(out)
	if err != nil {
		return backupRes, out, err
	}

	return backupRes, nil, nil
}

// newCommand initializes an instance of cli.CommandType with given parameters
func newCommand(command string, args ...string) cli.CommandType {
	// enable json-logging
	defaultArgs := []string{"--json"}
	return cli.CommandType{
		Binary:  binary,
		Command: command,
		Args:    append(defaultArgs, args...),
	}
}

// Ls executes "restic ls"
func Ls(ctx context.Context, glob *GlobalOptions, opts *LsOptions) ([]LsResult, error) {
	var args []string
	args = cli.StructToCLI(glob)
	args = append(args, cli.StructToCLI(opts)...)
	cmd := newCommand("ls", args...)

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	out = []byte(fmt.Sprint("[" +
		strings.ReplaceAll(strings.TrimRight(string(out), "\n"), "\n", ",") +
		"]"))
	var result []LsResult
	result, err = LsResponseFromJSON(out, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// LsResponseFromJSON unmarshals LS json output
func LsResponseFromJSON(jsonLog []byte, opts *LsOptions) ([]LsResult, error) {
	var result []LsResult
	var messages []LsMessage
	err := json.Unmarshal(jsonLog, &messages)
	if err != nil {
		return nil, err
	}
	var current LsResult
	for idx := range messages {
		currentMessage := messages[idx]
		if currentMessage.ShortID != nil {
			if *currentMessage.ShortID != "" {
				if current.SnapshotID != "" {
					result = append(result, current)
				}
				current = LsResult{
					SnapshotID: *currentMessage.ShortID,
					Paths:      currentMessage.Paths,
					Time:       currentMessage.Time,
					Files:      []LsFile{},
				}
				continue
			}
		}
		if !opts.Flags.Long {
			if currentMessage.Type == fileType {
				current.Files = append(current.Files, LsFile{
					Path: currentMessage.Path,
				})
			}
			continue
		}
		if currentMessage.Type == fileType {
			current.Size += currentMessage.Size
			// valid entries for files should have uid, guid and permissions if -l is set
			// use pointers to distinguish empty value from 0 for root user
			if currentMessage.Mode != nil && currentMessage.UID != nil && currentMessage.GID != nil {
				current.Files = append(current.Files, LsFile{
					Permissions: *currentMessage.Mode,
					User:        *currentMessage.UID,
					Group:       *currentMessage.GID,
					Size:        currentMessage.Size,
					Time:        currentMessage.Time,
					Path:        currentMessage.Path,
				})
			}
		}
	}

	result = append(result, current)
	return result, nil
}

// GetSnapshotSize returns the summed file size of the given snapshots in bytes, based of the "restic ls -l" command.
func GetSnapshotSize(ctx context.Context, snapshotIDs []string) (size uint64) {
	opts := StatsOptions{
		Flags: &StatsFlags{},
		IDs:   snapshotIDs,
	}
	cmd := newCommand("stats", cli.StructToCLI(&opts)...)

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return
	}

	var stats Stats
	if err = json.Unmarshal(out, &stats); err != nil {
		return
	}
	return stats.TotalSize
}

// GetSnapshotSizeByPath returns the summed file size (filtered by its path)...
// ...of the given snapshot id in bytes, based of the "restic ls -l" command.
func GetSnapshotSizeByPath(ctx context.Context, snapshotID, path string) (size uint64) {
	opts := LsOptions{
		Flags: &LsFlags{
			Long: true,
		},
		SnapshotIDs: []string{snapshotID},
	}
	ls, err := Ls(ctx, &GlobalOptions{}, &opts)

	if err != nil {
		return
	}

	for _, itm := range ls {
		for _, f := range itm.Files {
			if strings.HasPrefix(f.Path, path) {
				size += f.Size
			}
		}
	}
	return
}

// ListSnapshots executes "restic snapshots"
func ListSnapshots(ctx context.Context, opts *SnapshotOptions) ([]Snapshot, error) {
	cmd := newCommand("snapshots", cli.StructToCLI(&opts)...)

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var snapshots []Snapshot
	err = json.Unmarshal(out, &snapshots)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

// Find executes "restic find"
func Find(ctx context.Context, opts *FindOptions) ([]FindResult, error) {
	cmd := newCommand("find", cli.StructToCLI(&opts)...)
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var findResult []FindResult
	err = json.Unmarshal(out, &findResult)
	if err != nil {
		return nil, err
	}
	return findResult, nil
}

// Check executes "restic check"
func Check(ctx context.Context, flags *CheckFlags) ([]byte, error) {
	cmd := newCommand("check", cli.StructToCLI(flags)...)
	return cli.Run(ctx, cmd)
}

// Forget executes "restic forget"
func Forget(
	ctx context.Context, globalOpts *GlobalOptions, forgetOpts *ForgetOptions,
) (
	removedSnapshots []string, output []byte, err error,
) {
	forgetOpts.Flags.Compact = true // make sure compact mode is enabled to parse result correctly

	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(forgetOpts)...)

	cmd := cli.CommandType{
		Binary:  binary,
		Command: "forget",
		Args:    args,
	}

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, out, err
	}
	var deletedSnapshots []string
	var forgetResponse ForgetResponse
	err = json.Unmarshal(out, &forgetResponse)
	if err != nil {
		return nil, out, err
	}
	for idx := range forgetResponse.Tags {
		for index := range forgetResponse.Tags[idx].Remove {
			if forgetResponse.Tags[idx].Remove[index].ID != nil {
				deletedSnapshots = append(deletedSnapshots, *forgetResponse.Tags[idx].Remove[index].ID)
			}
		}
	}
	return deletedSnapshots, out, nil
}

// Prune executes "restic prune"
func Prune(ctx context.Context) ([]byte, error) {
	cmd := newCommand("prune", nil...)

	return cli.Run(ctx, cmd)
}

// RebuildIndex executes "restic rebuild-index"
func RebuildIndex(ctx context.Context) ([]byte, error) {
	nice := 19
	ionice := 2
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "rebuild-index",
		Args:    nil,
		Nice:    &nice,
		IONice:  &ionice,
	}
	return cli.Run(ctx, cmd)
}

// RestoreBackup executes "restic restore"
func RestoreBackup(ctx context.Context, glob *GlobalOptions, opts *RestoreOptions, unlock bool) ([]byte, error) {
	args := cli.StructToCLI(glob)
	args = append(args, cli.StructToCLI(opts)...)

	if unlock {
		unlockOpts := UnlockOptions{
			Flags: &UnlockFlags{
				RemoveAll: false,
			},
		}
		_, err := Unlock(ctx, glob, &unlockOpts)
		if err != nil {
			return nil, err
		}
	}

	cmd := newCommand("restore", args...)

	return cli.Run(ctx, cmd)
}

// Unlock executes "restic unlock"
func Unlock(ctx context.Context, globalOpts *GlobalOptions, unlockOpts *UnlockOptions) ([]byte, error) {
	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(unlockOpts)...)
	cmd := newCommand("unlock", args...)

	return cli.Run(ctx, cmd)
}

// Tag executes "restic tag"
func Tag(ctx context.Context, opts *TagOptions) ([]byte, error) {
	cmd := newCommand("tag", cli.StructToCLI(opts)...)

	return cli.Run(ctx, cmd)
}
