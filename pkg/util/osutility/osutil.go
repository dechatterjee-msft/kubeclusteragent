package osutility

type OSUtil interface {
	Exec() Exec
	Filesystem() Filesystem
	PackageManager() PackageManager
	Sysctl() Sysctl
	Systemd() Systemd
	Kubectl() Kubectl
	Kubeadm() Kubeadm
}

type DryRun struct {
	exec           *FakeExec
	filesystem     *FakeFilesystem
	packageManager *FakePackageManager
	sysctl         *FakeSysctl
	systemd        *FakeSystemd
	kubectl        *FakeKubectl
	kubeadm        *FakeKubeadm
}

var _ OSUtil = &DryRun{}

func NewDryRun() *DryRun {
	u := &DryRun{
		exec:           NewFakeExec(),
		filesystem:     NewFakeFilesystem(),
		packageManager: NewFakePackage(),
		sysctl:         NewFakeSysctl(),
		systemd:        NewFakeSystemd(),
		kubectl:        NewFakeKubectl(),
		kubeadm:        NewFakeKubeadm(),
	}
	return u
}

func (f *DryRun) PackageManager() PackageManager {
	return f.packageManager
}

func (f *DryRun) Kubectl() Kubectl {
	return f.kubectl
}

func (f *DryRun) Systemd() Systemd {
	return f.systemd
}

func (f *DryRun) Filesystem() Filesystem {
	return f.filesystem
}

func (f *DryRun) Sysctl() Sysctl {
	return f.sysctl
}

func (f *DryRun) Exec() Exec {
	return f.exec
}

func (f *DryRun) Kubeadm() Kubeadm {
	return f.kubeadm
}

type Live struct {
	exec           *LiveExec
	filesystem     *LiveFilesystem
	packageManager *LivePackageManager
	sysctl         *LiveSysctl
	systemd        *LiveSystemd
	kubectl        *LiveKubectl
	kubeadm        *LiveKubeadm
}

var _ OSUtil = &Live{}

func New() *Live {
	execUtil := NewLiveExec()
	fsUtil := NewLiveFilesystem()

	u := &Live{
		exec:           execUtil,
		filesystem:     fsUtil,
		packageManager: NewLivePackageManager(execUtil, fsUtil),
		sysctl:         NewLiveSysctl(execUtil, fsUtil),
		systemd:        NewLiveSystemd(execUtil),
		kubectl:        NewLiveKubectl(execUtil),
		kubeadm:        NewLiveKubeadm(execUtil),
	}

	return u
}

func (f *Live) Kubectl() Kubectl {
	return f.kubectl
}

func (f *Live) PackageManager() PackageManager {
	return f.packageManager
}

func (f *Live) Systemd() Systemd {
	return f.systemd
}

func (f *Live) Filesystem() Filesystem {
	return f.filesystem
}

func (f *Live) Sysctl() Sysctl {
	return f.sysctl
}

func (f *Live) Exec() Exec {
	return f.exec
}

func (f *Live) Kubeadm() Kubeadm {
	return f.kubeadm
}
