#include "perf_profiler.h"

typedef struct {
    void **address;
    size_t size;
    __u64 call_time;
} malloc_data_t;

typedef struct {
    __u64 pid;
    __u64 size;
    __u64 address;
    __u64 duration;
} meminfo_t;

MAP_HASH(cuda_malloc_info, __u64, malloc_data_t, 65536)
MAP_HASH(python_tstate_addr, __u32, __u64, 65536)

SEC("uretprobe/python_save_tstate")
int uprobe_python_save_tstate(struct pt_regs *ctx) {
    long ret = PT_REGS_RC(ctx);
    __u32 tgid = bpf_get_current_pid_tgid() >> 32;

	__u64 *addr = python_tstate_addr__lookup(&tgid);
    if (addr) {
        *addr = ret;
    } else {
        python_tstate_addr__update(&tgid, (__u64*) &ret);
    }
	return 0;
}

SEC("uprobe/cuda_malloc")
int uprobe_cuda_malloc(struct pt_regs *ctx) {
    void *address = (void *) PT_REGS_PARM1(ctx);
    size_t size = PT_REGS_PARM2(ctx);
    __u64 id = bpf_get_current_pid_tgid();
    malloc_data_t *data = cuda_malloc_info__lookup(&id);
    __u64 call_time = bpf_ktime_get_ns();
    if (data) {
        data->address = address;
        data->size = size;
        data->call_time = call_time;
    } else {
        malloc_data_t newdata = { .address = address, .size = size, .call_time = call_time };
        cuda_malloc_info__update(&id, &newdata);
    }

    void *allocated = NULL;
    bpf_probe_read_user(&allocated, sizeof(allocated), address);
    bpf_debug("pid %d tgid %d", id & 0xFFFFFFFF, id >> 32);
    bpf_debug("allocate %lu bytes to %lx@%lx", size, allocated, address);

    return 0;
}

SEC("uretprobe/cuda_malloc")
int uretprobe_cuda_malloc(struct pt_regs *ctx)
{
    long ret = PT_REGS_RC(ctx);
    if (ret != 0) {
        return 0;
    }

    __u64 id = bpf_get_current_pid_tgid();
    malloc_data_t *data = cuda_malloc_info__lookup(&id);
    if (!data) {
        return 0;
    }

    meminfo_t info = { .pid = id >> 32, .size = data->size, .duration = bpf_ktime_get_ns() - data->call_time };

    bpf_probe_read_user(&info.address, sizeof(info.address), data->address);
    bpf_debug("pid %d tgid %d", id & 0xFFFFFFFF, id >> 32);
    bpf_debug("allocated %lu to %lx@%lx", info.size, info.address, data->address);
    return 0;
}

#if 0
SEC("uprobe/cuda_memcpy_async")
int uprobe_cuda_memcpy_async(struct pt_regs *ctx)
{
    int kind = PT_REGS_PARM4(ctx);
    size_t count = PT_REGS_PARM5(ctx);

    bpf_debug("cudaMemcpyAsync copy %d %d bytes", kind, count);
    return 0;
}
#endif
