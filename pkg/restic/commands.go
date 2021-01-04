package restic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mittwald/brudi/pkg/cli"
)

const (
	binary   = "restic"
	fileType = "file"
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
func parseSnapshotOut(responses []map[string]interface{}) (BackupResult, error) {
	var result BackupResult

	var parentSnapshotID string
	var snapshotID string
	for idx := range responses {
		v := responses[idx]
		if v["message_type"] == "summary" {
			if v["snapshot_id"] != nil {
				snapshotID = v["snapshot_id"].(string)
			}
			if v["parent"] != nil {
				parentSnapshotID = v["parent"].(string)
			}
		}
	}
	if parentSnapshotID != "" {
		result.ParentSnapshotID = parentSnapshotID
	}
	if snapshotID == "" {
		return BackupResult{}, fmt.Errorf("failed to parse snapshotID")
	}
	result.SnapshotID = snapshotID
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
		strings.Replace(strings.TrimRight(string(out), "\n"), "\n", ",", -1) +
		"]"))
	responseList := []map[string]interface{}{}
	jerr := json.Unmarshal(out, &responseList)
	if jerr != nil {
		fmt.Println(jerr)
	}

	backupRes, err := parseSnapshotOut(responseList)
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
func Ls(ctx context.Context, opts *LsOptions) ([]LsResult, error) {
	var result []LsResult
	cmd := newCommand("ls", cli.StructToCLI(&opts)...)

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(out)
	bufReader := bufio.NewReader(reader)
	var current LsResult
	line, _, err := bufReader.ReadLine()

	for err != nil {
		var mess LsMessage
		jerr := json.Unmarshal(line, &mess)
		if jerr != nil {
			return nil, jerr
		}
		if mess.ShortID != "" {
			if current.SnapshotID != "" {
				result = append(result, current)
			}
			current = LsResult{
				SnapshotID: mess.ShortID,
				Paths:      mess.Paths,
				Time:       mess.Time,
				Files:      []LsFile{},
			}
			line, _, err = bufReader.ReadLine()
			continue
		}

		if !opts.Flags.Long {
			if mess.Type == fileType {
				current.Files = append(current.Files, LsFile{
					Path: mess.Path,
				})
				line, _, err = bufReader.ReadLine()
				continue
			}
		}
		if mess.Type == fileType {
			current.Size += mess.Size
			current.Files = append(current.Files, LsFile{
				Permissions: mess.Mode,
				User:        mess.UID,
				Group:       mess.GID,
				Size:        mess.Size,
				Time:        mess.Time,
				Path:        mess.Path,
			})
		}
		line, _, err = bufReader.ReadLine()
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
	ls, err := Ls(ctx, &opts)

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
	if err := json.Unmarshal(out, &snapshots); err != nil {
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
	if err := json.Unmarshal(out, &findResult); err != nil {
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
			deletedSnapshots = append(deletedSnapshots, forgetResponse.Tags[idx].Remove[index].ID)
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
func RestoreBackup(ctx context.Context, opts *RestoreOptions) ([]byte, error) {
	cmd := newCommand("restore", cli.StructToCLI(opts)...)

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
