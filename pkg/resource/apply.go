package resource

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	kresource "k8s.io/cli-runtime/pkg/resource"
	kcmdapply "k8s.io/kubectl/pkg/cmd/apply"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// ApplyResult indicates what type of change was performed
// by calling the Apply function
type ApplyResult string

const (
	// ConfiguredApplyResult is returned when a patch was submitted
	ConfiguredApplyResult ApplyResult = "configured"

	// UnchangedApplyResult is returned when no change occurred
	UnchangedApplyResult ApplyResult = "unchanged"

	// CreatedApplyResult is returned when a resource was created
	CreatedApplyResult ApplyResult = "created"

	// UnknownApplyResult is returned when the resulting action could not be determined
	UnknownApplyResult ApplyResult = "unknown"
)

const fieldTooLong metav1.CauseType = "FieldValueTooLong"

// Apply applies the given resource bytes to the target cluster specified by kubeconfig
func (r *helper) Apply(obj []byte) (ApplyResult, error) {
	factory, err := r.getFactory("")
	if err != nil {
		r.logger.WithError(err).Error("failed to obtain factory for apply")
		return "", err
	}
	ioStreams := genericclioptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	applyOptions, changeTracker, err := r.setupApplyCommand(factory, obj, ioStreams)
	if err != nil {
		r.logger.WithError(err).Error("failed to setup apply command")
		return "", err
	}
	if applyOptions == nil {
		r.logger.WithError(err).Error("failed to setup applyOptions")
	}
	if changeTracker == nil {
		r.logger.WithError(err).Error("failed to setup apply command")
	}

	r.mockApply()

	return ConfiguredApplyResult, nil
}

// ApplyRuntimeObject serializes an object and applies it to the target cluster specified by the kubeconfig.
func (r *helper) ApplyRuntimeObject(obj runtime.Object, scheme *runtime.Scheme) (ApplyResult, error) {
	data, err := Serialize(obj, scheme)
	if err != nil {
		r.logger.WithError(err).Warn("cannot serialize runtime object")
		return "", err
	}
	return r.Apply(data)
}

func (r *helper) CreateOrUpdate(obj []byte) (ApplyResult, error) {
	factory, err := r.getFactory("")
	if err != nil {
		r.logger.WithError(err).Error("failed to obtain factory for apply")
		return "", err
	}

	errOut := &bytes.Buffer{}
	result, err := r.createOrUpdate(factory, obj, errOut)
	if err != nil {
		r.logger.WithError(err).
			WithField("stderr", errOut.String()).Warn("running the apply command failed")
		return "", err
	}
	return result, nil
}

func (r *helper) CreateOrUpdateRuntimeObject(obj runtime.Object, scheme *runtime.Scheme) (ApplyResult, error) {
	data, err := Serialize(obj, scheme)
	if err != nil {
		r.logger.WithError(err).Warn("cannot serialize runtime object")
		return "", err
	}
	return r.CreateOrUpdate(data)
}

func (r *helper) Create(obj []byte) (ApplyResult, error) {
	factory, err := r.getFactory("")
	if err != nil {
		r.logger.WithError(err).Error("failed to obtain factory for apply")
		return "", err
	}
	result, err := r.createOnly(factory, obj)
	if err != nil {
		r.logger.WithError(err).Warn("running the create command failed")
		return "", err
	}
	return result, nil
}

func (r *helper) CreateRuntimeObject(obj runtime.Object, scheme *runtime.Scheme) (ApplyResult, error) {
	data, err := Serialize(obj, scheme)
	if err != nil {
		r.logger.WithError(err).Warn("cannot serialize runtime object")
		return "", err
	}
	return r.Create(data)
}

func (r *helper) mockApply() {
	// do a GET to http://httpwait-default.apps.gshereme-spoke-1.new-installer.openshift.com/?wait=3000
	// use a random wait time
	in := []int{100, 200, 300, 100, 200, 300, 100, 200, 300, 100, 200, 300, 500, 500, 3000}
	randomIndex := rand.Intn(len(in))
	wait := in[randomIndex]

	r.logger.Debug("running fake apply for ", wait, "ms")
	url := fmt.Sprintf("http://httpwait-default.apps.gshereme-spoke-1.new-installer.openshift.com/?wait=%d", wait)
	r.logger.Debug("getting ", url)
	resp, err := http.Get(url)
	if err != nil {
		r.logger.Debug("err: ", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("Response status: ", resp.Status)
}

func (r *helper) createOnly(f cmdutil.Factory, obj []byte) (ApplyResult, error) {
	info, err := r.getResourceInternalInfo(f, obj)
	if err != nil {
		return "", err
	}
	if info == nil {
		r.logger.Debug("err getting info")
	}
	c, err := f.DynamicClient()
	if err != nil {
		r.logger.Debug("err getting client")
	}
	if c == nil {
		r.logger.Debug("err getting client")
	}

	r.mockApply()

	return CreatedApplyResult, nil
}

func (r *helper) createOrUpdate(f cmdutil.Factory, obj []byte, errOut io.Writer) (ApplyResult, error) {
	info, err := r.getResourceInternalInfo(f, obj)
	if err != nil {
		return "", err
	}
	client, err := f.DynamicClient()
	if err != nil {
		r.logger.Debug("err getting client")
	}
	if client == nil {
		r.logger.Debug("err getting client")
	}
	sourceObj := info.Object.DeepCopyObject()
	if sourceObj == nil {
		r.logger.Debug("err getting sourceObj")
	}

	r.mockApply()

	return CreatedApplyResult, nil
}

func (r *helper) setupApplyCommand(f cmdutil.Factory, obj []byte, ioStreams genericclioptions.IOStreams) (*kcmdapply.ApplyOptions, *changeTracker, error) {
	r.logger.Debug("setting up apply command")
	o := kcmdapply.NewApplyOptions(ioStreams)
	dynamicClient, err := f.DynamicClient()
	if err != nil {
		r.logger.WithError(err).Error("cannot obtain dynamic client from factory")
		return nil, nil, err
	}
	o.DeleteOptions = o.DeleteFlags.ToOptions(dynamicClient, o.IOStreams)
	o.OpenAPISchema, _ = f.OpenAPISchema()
	o.Validator, err = f.Validator(false)
	if err != nil {
		r.logger.WithError(err).Error("cannot obtain schema to validate objects from factory")
		return nil, nil, err
	}
	o.Builder = f.NewBuilder()
	o.Mapper, err = f.ToRESTMapper()
	if err != nil {
		r.logger.WithError(err).Error("cannot obtain RESTMapper from factory")
		return nil, nil, err
	}

	o.DynamicClient = dynamicClient
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		r.logger.WithError(err).Error("cannot obtain namespace from factory")
		return nil, nil, err
	}
	tracker := &changeTracker{
		internalToPrinter: func(string) (printers.ResourcePrinter, error) { return o.PrintFlags.ToPrinter() },
	}
	o.ToPrinter = tracker.ToPrinter
	info, err := r.getResourceInternalInfo(f, obj)
	if err != nil {
		return nil, nil, err
	}
	o.SetObjects([]*kresource.Info{info})
	return o, tracker, nil
}

type trackerPrinter struct {
	setResult       func()
	internalPrinter printers.ResourcePrinter
}

func (p *trackerPrinter) PrintObj(o runtime.Object, w io.Writer) error {
	if p.setResult != nil {
		p.setResult()
	}
	return p.internalPrinter.PrintObj(o, w)
}

type changeTracker struct {
	result            []ApplyResult
	internalToPrinter func(string) (printers.ResourcePrinter, error)
}

func (t *changeTracker) GetResult() ApplyResult {
	if len(t.result) == 1 {
		return t.result[0]
	}
	return UnknownApplyResult
}

func (t *changeTracker) ToPrinter(name string) (printers.ResourcePrinter, error) {
	var f func()
	switch name {
	case "created":
		f = func() { t.result = append(t.result, CreatedApplyResult) }
	case "configured":
		f = func() { t.result = append(t.result, ConfiguredApplyResult) }
	case "unchanged":
		f = func() { t.result = append(t.result, UnchangedApplyResult) }
	}
	p, err := t.internalToPrinter(name)
	if err != nil {
		return nil, err
	}
	return &trackerPrinter{
		internalPrinter: p,
		setResult:       f,
	}, nil
}
