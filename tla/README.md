# TLA+ Specification and Trace Validation for Raft Library: A Comprehensive Guide

This document presents comprehensive guidelines on how to create a TLA+ specification for the Raft library as implemented in `etcd-io/raft`. The distinctive behaviors of this library, such as reconfiguration, set it apart from the original Raft algorithm. A TLA+ specification serves dual purposes: it not only helps verify the correctness of the model but also facilitates model-based trace validation.

## Checking the Model with TLA+ Specifications

The TLA+ specifications can be verified using the TLC model checker. Here are a few methods:

1. **TLA+ Toolbox**: This is an ideal tool for an in-depth study of the specification. For more information, refer to the [TLA+ Toolbox guideline](https://lamport.azurewebsites.net/tla/toolbox.html).

2. **VSCode Plugin TLA+ Nightly**: This is a viable alternative to the TLA+ Toolbox, particularly for those accustomed to using VSCode. For more information, refer to the [VSCode Plugin TLA+ Nightly guideline](https://github.com/tlaplus/vscode-tlaplus/wiki).

3. **CLI**: This is the best option for integration with existing test frameworks and automation. For more information, refer to the [CLI guideline](https://learntla.com/topics/cli.html). You can execute a typical command line to verify etcdraft.tla as shown below. Please ensure that tla2tools.jar and CommunityModules-deps.jar have been downloaded and are available in the current folder before proceeding.

    ```console
    java -XX:+UseParallelGC -Dtlc2.tool.impl.Tool.cdot=true -cp tla2tools.jar:CommunityModules-deps.jar tlc2.TLC -workers auto MCetcdraft.tla -dumpTrace tla MCetcdraft.trace.tla -dumpTrace json MCetcdraft.json -lncheck final
    ```

## TLA+ Model-Based Trace Validation

The correctness of applications based on a consensus algorithm is ensured by two factors: the correctness of the algorithm itself, and the alignment of the implementation with the algorithm. 

The first factor, the correctness of the algorithm, is assured through model checking the specification. The second component, the alignment between the implementation and the algorithm, is fortified through trace validation, which serves to bridge the gap between the model and its implementation.

To ensure the robustness of trace validation, we initially inject multiple trace log points at specific sections in the code where state transitions occur (for example, SendAppendEntries, BecomeLeader, etcd). The trace validation specification leverages these traces as a state space constraint, guiding the state machine to traverse in a manner identical to that of the service.

If a trace suggests a state or transition that the state machine can't accommodate, it indicates a discrepancy between the model and its implementation. In cases where the model has already been verified by the TLC model checker, it's more likely that any issues arise from the implementation rather than the model.

## Enabling and Running TLA+ Trace Validation

To activate trace collection, build the application using the "with_tla" tag (`go build -tags=with_tla`). The `StartNode` and/or `RestartNode` should be invoked with `Config` and the `TraceLogger` property set to an instance of `TraceLogger` interface. This instance emits tracing events in the correct format.

Here is any example trace logger using zap.Logger (sampling shall be disabled to ensure all traces are logged)

```go
type TraceLogger interface {
	TraceEvent(*TracingEvent)
}

type MyTraceLogger struct{
  lg *zap.Logger
}

func (t *MyTraceLogger) TraceEvent(e *TracingEvent) {
  t.lg.Debug("trace", zap.String("tag", "trace"), zap.Any("event", ev))
}

```

To start a three-node cluster with trace logger
```go
storage := raft.NewMemoryStorage()
  c := &raft.Config{
    ID:              0x01,
    ElectionTick:    10,
    HeartbeatTick:   1,
    Storage:         storage,
    MaxSizePerMsg:   4096,
    MaxInflightMsgs: 256,
    TraceLogger:     &MyTraceLogger{},
  }
  // Set peer list to the other nodes in the cluster.
  // Note that they need to be started separately as well.
  n := raft.StartNode(c, []raft.Peer{{ID: 0x02}, {ID: 0x03}})
```
To preserve the causality of events across nodes, run all application instances on the same machine and store traces in the same file. This approach maintains the order of traces in all instances.

With above example trace logger, validate.sh can be used to validate traces parallelly.
```console
./validate.sh -s ./Traceetcdraft.tla -c ./Traceetcdraft.cfg /tmp/ramdisk/*.ndjson
```
**Note**： Trace validation can take very long time if a trace file contains too many log lines. Environment variable MAX_TRACE can be set to only investigate top N log lines in each trace file.
