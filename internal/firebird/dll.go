package firebird

import (
	"fmt"
	"syscall"
	"unsafe"
)

type FBClient struct {
	dll                   *syscall.LazyDLL
	AttachDatabase        *syscall.LazyProc
	DetachDatabase        *syscall.LazyProc
	StartTransaction      *syscall.LazyProc
	CommitTransaction     *syscall.LazyProc
	RollbackTransaction   *syscall.LazyProc
	DsqlAllocateStatement *syscall.LazyProc
	DsqlPrepare           *syscall.LazyProc
	DsqlDescribe          *syscall.LazyProc
	DsqlDescribeBind      *syscall.LazyProc
	DsqlExecute           *syscall.LazyProc
	DsqlFetch             *syscall.LazyProc
	DsqlFreeStatement     *syscall.LazyProc
	OpenBlob              *syscall.LazyProc
	GetSegment            *syscall.LazyProc
	CloseBlob             *syscall.LazyProc
	Interprete            *syscall.LazyProc
}

func LoadFBClient(dllDir string) (*FBClient, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setDllDir := kernel32.NewProc("SetDllDirectoryW")
	dirPtr, _ := syscall.UTF16PtrFromString(dllDir)
	setDllDir.Call(uintptr(unsafe.Pointer(dirPtr)))

	dll := syscall.NewLazyDLL(dllDir + "\\fbclient.dll")
	err := dll.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load fbclient.dll: %w", err)
	}

	return &FBClient{
		dll:                   dll,
		AttachDatabase:        dll.NewProc("isc_attach_database"),
		DetachDatabase:        dll.NewProc("isc_detach_database"),
		StartTransaction:      dll.NewProc("isc_start_transaction"),
		CommitTransaction:     dll.NewProc("isc_commit_transaction"),
		RollbackTransaction:   dll.NewProc("isc_rollback_transaction"),
		DsqlAllocateStatement: dll.NewProc("isc_dsql_allocate_statement"),
		DsqlPrepare:           dll.NewProc("isc_dsql_prepare"),
		DsqlDescribe:          dll.NewProc("isc_dsql_describe"),
		DsqlDescribeBind:      dll.NewProc("isc_dsql_describe_bind"),
		DsqlExecute:           dll.NewProc("isc_dsql_execute"),
		DsqlFetch:             dll.NewProc("isc_dsql_fetch"),
		DsqlFreeStatement:     dll.NewProc("isc_dsql_free_statement"),
		OpenBlob:              dll.NewProc("isc_open_blob2"),
		GetSegment:            dll.NewProc("isc_get_segment"),
		CloseBlob:             dll.NewProc("isc_close_blob"),
		Interprete:            dll.NewProc("isc_interprete"),
	}, nil
}
