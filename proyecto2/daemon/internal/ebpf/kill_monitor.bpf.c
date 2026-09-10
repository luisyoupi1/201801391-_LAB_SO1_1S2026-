//go:build ignore

// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>

struct kill_event {
    __u32 source_pid;
    __u32 target_pid;
    __s32 signal;
    __u32 padding;
    __u64 timestamp_ns;
    char command[16];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 20);
} events SEC(".maps");

const struct kill_event *unused_event __attribute__((unused));

SEC("tracepoint/syscalls/sys_enter_kill")
int trace_kill(struct trace_event_raw_sys_enter *ctx)
{
    struct kill_event *event;
    __u64 pid_tgid;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    pid_tgid = bpf_get_current_pid_tgid();
    event->source_pid = pid_tgid >> 32;
    event->target_pid = (__u32)ctx->args[0];
    event->signal = (__s32)ctx->args[1];
    event->padding = 0;
    event->timestamp_ns = bpf_ktime_get_ns();
    bpf_get_current_comm(&event->command, sizeof(event->command));
    bpf_ringbuf_submit(event, 0);
    return 0;
}

/* Observe accepted termination signals independently of the sending syscall.
 * task->tgid is the host process ID, including senders in PID namespaces.
 */
SEC("raw_tracepoint/signal_generate")
int trace_signal(struct bpf_raw_tracepoint_args *ctx)
{
    int sig = (int)ctx->args[0];
    struct task_struct *task = (struct task_struct *)ctx->args[2];
    int result = (int)ctx->args[4];
    struct kill_event *event;

    if ((sig != 9 && sig != 15) || result != 0)
        return 0;
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;
    event->source_pid = bpf_get_current_pid_tgid() >> 32;
    event->target_pid = BPF_CORE_READ(task, tgid);
    event->signal = sig;
    event->padding = 1; /* signal_generate, distinct from syscall observations */
    event->timestamp_ns = bpf_ktime_get_ns();
    bpf_get_current_comm(&event->command, sizeof(event->command));
    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";

