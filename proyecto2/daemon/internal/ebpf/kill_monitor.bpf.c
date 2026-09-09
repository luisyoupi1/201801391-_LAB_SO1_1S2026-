// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

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

char LICENSE[] SEC("license") = "Dual BSD/GPL";

