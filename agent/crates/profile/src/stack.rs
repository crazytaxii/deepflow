use std::ffi::CStr;
use std::io::Write;
use std::mem;

const PYEVAL_FNAME: &'static str = "_PyEval_EvalFrameDefault";

#[no_mangle]
pub unsafe extern "C" fn merge_stacks(
    trace_str: *mut i8,
    len: usize,
    i_trace: *const i8,
    u_trace: *const i8,
) {
    let Ok(i_trace) = CStr::from_ptr(i_trace).to_str() else {
        return;
    };
    let Ok(u_trace) = CStr::from_ptr(u_trace).to_str() else {
        return;
    };
    trace_str.write_bytes(0, len);
    let mut trace_str = Vec::from_raw_parts(trace_str as *mut u8, 0, len);

    let n_py_frames = i_trace.split(";").count() - 1; // <module> does not count
    let n_eval_frames = u_trace
        .split(";")
        .filter(|c_func| *c_func == PYEVAL_FNAME)
        .count();

    if n_eval_frames == 0 {
        // native stack not correctly unwinded, just put it on top of python frames
        let _ = write!(
            &mut trace_str,
            "{};[lost] incomplete python c stack;{}",
            i_trace, u_trace
        );
    } else if n_py_frames == n_eval_frames {
        // no native stack
        let _ = write!(&mut trace_str, "{}", i_trace);
    } else if n_py_frames == n_eval_frames - 1 {
        // python calls native, just put everything after the last _PyEval on top of python frames (including the semicolon)
        let loc = u_trace.rfind(PYEVAL_FNAME).unwrap() + PYEVAL_FNAME.len();
        let _ = write!(&mut trace_str, "{}{}", i_trace, &u_trace[loc..]);
    } else {
        println!("u_trace: {}\ni_trace: {}\n", u_trace, i_trace);
    }

    mem::forget(trace_str);
}
