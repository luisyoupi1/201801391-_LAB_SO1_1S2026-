// SPDX-License-Identifier: GPL-2.0
#include <linux/init.h>
#include <linux/kernel.h>
#include <linux/ktime.h>
#include <linux/math64.h>
#include <linux/mm.h>
#include <linux/module.h>
#include <linux/pid.h>
#include <linux/proc_fs.h>
#include <linux/rcupdate.h>
#include <linux/sched/cputime.h>
#include <linux/sched/mm.h>
#include <linux/sched/signal.h>
#include <linux/seq_file.h>
#include <linux/slab.h>
#include <linux/sysinfo.h>
#include <linux/vmalloc.h>

#ifndef CARNET
#define CARNET "201801391"
#endif

#define MODULE_NAME "continfo_pr2_so1"
#define PROC_NAME "continfo_pr2_so1_" CARNET
#define MAX_CAPTURED_PIDS 65536
#define CMDLINE_SIZE 512

static struct proc_dir_entry *proc_entry;

static unsigned long pages_to_kb(unsigned long pages)
{
    return pages << (PAGE_SHIFT - 10);
}

static void seq_put_json_string(struct seq_file *m, const char *value, size_t len)
{
    size_t i;

    seq_putc(m, '"');
    for (i = 0; i < len && value[i] != '\0'; i++) {
        unsigned char ch = value[i];

        switch (ch) {
        case '"':
            seq_puts(m, "\\\"");
            break;
        case '\\':
            seq_puts(m, "\\\\");
            break;
        case '\n':
            seq_puts(m, "\\n");
            break;
        case '\r':
            seq_puts(m, "\\r");
            break;
        case '\t':
            seq_puts(m, "\\t");
            break;
        default:
            if (ch < 0x20)
                seq_printf(m, "\\u%04x", ch);
            else
                seq_putc(m, ch);
        }
    }
    seq_putc(m, '"');
}

static int collect_pids(pid_t **result)
{
    struct task_struct *task;
    pid_t *pids;
    int count = 0;

    pids = kvmalloc_array(MAX_CAPTURED_PIDS, sizeof(*pids), GFP_KERNEL);
    if (!pids)
        return -ENOMEM;

    rcu_read_lock();
    for_each_process(task) {
        if (count >= MAX_CAPTURED_PIDS)
            break;
        pids[count++] = task_pid_nr(task);
    }
    rcu_read_unlock();

    *result = pids;
    return count;
}

static void normalize_cmdline(char *cmdline, int length)
{
    int i;

    for (i = 0; i < length; i++) {
        if (cmdline[i] == '\0')
            cmdline[i] = ' ';
    }
    while (length > 0 && cmdline[length - 1] == ' ')
        cmdline[--length] = '\0';
}

static int read_task_cmdline(struct task_struct *task, char *buffer,
                             size_t buffer_size)
{
    struct mm_struct *mm;
    unsigned long arg_start, arg_end;
    size_t length;
    int copied;

    if (!buffer_size)
        return 0;

    mm = get_task_mm(task);
    if (!mm)
        return 0;

    mmap_read_lock(mm);
    arg_start = mm->arg_start;
    arg_end = mm->arg_end;
    mmap_read_unlock(mm);

    if (arg_end <= arg_start) {
        mmput(mm);
        return 0;
    }

    length = min_t(size_t, arg_end - arg_start, buffer_size - 1);
    copied = access_process_vm(task, arg_start, buffer, length, FOLL_FORCE);
    mmput(mm);
    if (copied <= 0)
        return 0;

    buffer[copied] = '\0';
    return copied;
}

static void emit_process(struct seq_file *m, pid_t nr, unsigned long total_kb,
                         bool *first)
{
    struct task_struct *task;
    struct mm_struct *mm;
    struct pid *pid;
    unsigned long vsz_kb = 0;
    unsigned long rss_kb = 0;
    u64 utime = 0, stime = 0, elapsed_ns, cpu_basis_points = 0;
    u64 memory_basis_points = 0;
    char comm[TASK_COMM_LEN];
    char *cmdline;
    int cmdline_len;

    pid = find_get_pid(nr);
    if (!pid)
        return;
    task = get_pid_task(pid, PIDTYPE_PID);
    put_pid(pid);
    if (!task)
        return;

    cmdline = kzalloc(CMDLINE_SIZE, GFP_KERNEL);
    if (!cmdline) {
        put_task_struct(task);
        return;
    }

    get_task_comm(comm, task);
    cmdline_len = read_task_cmdline(task, cmdline, CMDLINE_SIZE);
    normalize_cmdline(cmdline, cmdline_len);

    mm = get_task_mm(task);
    if (mm) {
        vsz_kb = pages_to_kb(mm->total_vm);
        rss_kb = pages_to_kb(get_mm_rss(mm));
        mmput(mm);
    }

    task_cputime_adjusted(task, &utime, &stime);
    elapsed_ns = ktime_get_boottime_ns();
    if (elapsed_ns > task->start_boottime)
        elapsed_ns -= task->start_boottime;
    else
        elapsed_ns = 0;
    if (elapsed_ns)
        cpu_basis_points = div64_u64((utime + stime) * 10000ULL, elapsed_ns);
    if (total_kb)
        memory_basis_points = div64_u64((u64)rss_kb * 10000ULL, total_kb);

    if (!*first)
        seq_puts(m, ",\n");
    *first = false;

    seq_printf(m, "    {\"pid\":%d,\"name\":", nr);
    seq_put_json_string(m, comm, strnlen(comm, TASK_COMM_LEN));
    seq_puts(m, ",\"cmdline\":");
    if (cmdline_len > 0)
        seq_put_json_string(m, cmdline, cmdline_len);
    else
        seq_put_json_string(m, comm, strnlen(comm, TASK_COMM_LEN));
    seq_printf(m,
               ",\"vsz_kb\":%lu,\"rss_kb\":%lu,"
               "\"memory_percent\":%llu.%02llu,"
               "\"cpu_percent\":%llu.%02llu}",
               vsz_kb, rss_kb,
               (unsigned long long)(memory_basis_points / 100),
               (unsigned long long)(memory_basis_points % 100),
               (unsigned long long)(cpu_basis_points / 100),
               (unsigned long long)(cpu_basis_points % 100));

    kfree(cmdline);
    put_task_struct(task);
}

static int continfo_show(struct seq_file *m, void *v)
{
    struct sysinfo info;
    unsigned long total_kb, free_kb, used_kb;
    pid_t *pids = NULL;
    bool first = true;
    int count, i;

    (void)v;
    si_meminfo(&info);
    total_kb = pages_to_kb(info.totalram);
    free_kb = pages_to_kb(info.freeram);
    used_kb = total_kb >= free_kb ? total_kb - free_kb : 0;

    seq_printf(m,
               "{\n  \"carnet\":\"%s\",\n"
               "  \"memory\":{\"total_kb\":%lu,\"free_kb\":%lu,"
               "\"used_kb\":%lu},\n  \"processes\":[\n",
               CARNET, total_kb, free_kb, used_kb);

    count = collect_pids(&pids);
    if (count > 0) {
        for (i = 0; i < count; i++)
            emit_process(m, pids[i], total_kb, &first);
    }
    kvfree(pids);
    seq_puts(m, "\n  ]\n}\n");
    return 0;
}

static int continfo_open(struct inode *inode, struct file *file)
{
    return single_open(file, continfo_show, NULL);
}

static const struct proc_ops continfo_ops = {
    .proc_open = continfo_open,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int __init continfo_init(void)
{
    proc_entry = proc_create(PROC_NAME, 0444, NULL, &continfo_ops);
    if (!proc_entry)
        return -ENOMEM;
    pr_info("%s: /proc/%s creado para carnet %s\n",
            MODULE_NAME, PROC_NAME, CARNET);
    return 0;
}

static void __exit continfo_exit(void)
{
    proc_remove(proc_entry);
    pr_info("%s: /proc/%s eliminado\n", MODULE_NAME, PROC_NAME);
}

module_init(continfo_init);
module_exit(continfo_exit);

MODULE_LICENSE("GPL");
MODULE_AUTHOR("201801391");
MODULE_DESCRIPTION("Telemetria JSON de memoria y procesos para SO1 Proyecto 2");
MODULE_VERSION("1.0.0");
