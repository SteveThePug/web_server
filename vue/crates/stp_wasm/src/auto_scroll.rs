use std::cell::RefCell;
use std::rc::Rc;
use wasm_bindgen::prelude::*;
use web_sys::HtmlElement;

const SPEED: f64 = 0.0005; // % per frame
const PAUSE: i32 = 2000; // ms at top/bottom

struct Inner {
    el: HtmlElement,
    pos: f64,
    direction: f64, // 1.0 = down, -1.0 = up
    hovered: bool,
    cached_scroll_height: f64,
    raf_id: Option<i32>,
    pause_timeout_id: Option<i32>,
    resize_observer: Option<web_sys::ResizeObserver>,
    // closures kept alive
    tick_closure: Option<Closure<dyn FnMut()>>,
    resize_closure: Option<Closure<dyn FnMut(js_sys::Array)>>,
    mouseenter_closure: Option<Closure<dyn FnMut()>>,
    mouseleave_closure: Option<Closure<dyn FnMut()>>,
    start_after_pause_closure: Option<Closure<dyn FnMut()>>,
}

#[wasm_bindgen]
pub struct AutoScroller {
    inner: Rc<RefCell<Inner>>,
}

impl Inner {
    fn measure_scroll_height(&mut self) {
        self.cached_scroll_height = self.el.scroll_height() as f64;
    }

    fn stop_loop(&mut self) {
        let window = web_sys::window().unwrap();
        if let Some(id) = self.raf_id.take() {
            window.cancel_animation_frame(id).ok();
        }
        if let Some(id) = self.pause_timeout_id.take() {
            window.clear_timeout_with_handle(id);
        }
    }
}

fn start_loop(inner: &Rc<RefCell<Inner>>) {
    {
        let mut s = inner.borrow_mut();
        s.stop_loop();
    }
    schedule_tick(inner);
}

fn schedule_tick(inner: &Rc<RefCell<Inner>>) {
    let inner_clone = Rc::clone(inner);

    // Create a fresh tick closure each frame
    let closure = Closure::once(move || {
        tick(&inner_clone);
    });

    let mut s = inner.borrow_mut();
    let window = web_sys::window().unwrap();
    let id = window
        .request_animation_frame(closure.as_ref().unchecked_ref())
        .unwrap();
    s.raf_id = Some(id);
    // Store the closure to keep it alive until the frame fires
    s.tick_closure = Some(closure);
}

fn tick(inner: &Rc<RefCell<Inner>>) {
    let should_continue;
    {
        let mut s = inner.borrow_mut();
        s.raf_id = None;

        if s.hovered {
            return;
        }

        if s.cached_scroll_height == 0.0 {
            drop(s);
            schedule_tick(inner);
            return;
        }

        let reached_bottom = s.pos >= 1.0;
        let reached_top = s.pos <= 0.0;

        if reached_bottom {
            s.pos = 0.999;
            s.direction = -1.0;
            drop(s);
            schedule_pause(inner);
            return;
        } else if reached_top && s.direction == -1.0 {
            s.pos = 0.001;
            s.direction = 1.0;
            drop(s);
            schedule_pause(inner);
            return;
        }

        s.pos += s.direction * SPEED;
        s.el.set_scroll_top((s.pos * s.cached_scroll_height) as i32);
        should_continue = true;
    }

    if should_continue {
        schedule_tick(inner);
    }
}

fn schedule_pause(inner: &Rc<RefCell<Inner>>) {
    {
        let mut s = inner.borrow_mut();
        s.stop_loop();
    }

    let inner_clone = Rc::clone(inner);
    let closure = Closure::once(move || {
        start_loop(&inner_clone);
    });

    let mut s = inner.borrow_mut();
    let window = web_sys::window().unwrap();
    let id = window
        .set_timeout_with_callback_and_timeout_and_arguments_0(
            closure.as_ref().unchecked_ref(),
            PAUSE,
        )
        .unwrap();
    s.pause_timeout_id = Some(id);
    s.start_after_pause_closure = Some(closure);
}

#[wasm_bindgen]
impl AutoScroller {
    #[wasm_bindgen(constructor)]
    pub fn new(el: HtmlElement) -> AutoScroller {
        let inner = Rc::new(RefCell::new(Inner {
            el,
            pos: 0.0,
            direction: 1.0,
            hovered: false,
            cached_scroll_height: 0.0,
            raf_id: None,
            pause_timeout_id: None,
            resize_observer: None,
            tick_closure: None,
            resize_closure: None,
            mouseenter_closure: None,
            mouseleave_closure: None,
            start_after_pause_closure: None,
        }));

        // Set up mouseenter listener
        {
            let inner_clone = Rc::clone(&inner);
            let closure = Closure::wrap(Box::new(move || {
                let mut s = inner_clone.borrow_mut();
                s.hovered = true;
                s.stop_loop();
            }) as Box<dyn FnMut()>);
            let s = inner.borrow();
            s.el
                .add_event_listener_with_callback("mouseenter", closure.as_ref().unchecked_ref())
                .unwrap();
            drop(s);
            inner.borrow_mut().mouseenter_closure = Some(closure);
        }

        // Set up mouseleave listener
        {
            let inner_clone = Rc::clone(&inner);
            let closure = Closure::wrap(Box::new(move || {
                {
                    let mut s = inner_clone.borrow_mut();
                    s.hovered = false;
                    if s.cached_scroll_height > 0.0 {
                        s.pos = s.el.scroll_top() as f64 / s.cached_scroll_height;
                    }
                }
                start_loop(&inner_clone);
            }) as Box<dyn FnMut()>);
            let s = inner.borrow();
            s.el
                .add_event_listener_with_callback("mouseleave", closure.as_ref().unchecked_ref())
                .unwrap();
            drop(s);
            inner.borrow_mut().mouseleave_closure = Some(closure);
        }

        AutoScroller { inner }
    }

    pub fn start(&self) {
        // Measure initial scroll height
        self.inner.borrow_mut().measure_scroll_height();

        // Set up resize observer
        let inner_clone = Rc::clone(&self.inner);
        let resize_closure = Closure::wrap(Box::new(move |_entries: js_sys::Array| {
            inner_clone.borrow_mut().measure_scroll_height();
        }) as Box<dyn FnMut(js_sys::Array)>);

        let observer =
            web_sys::ResizeObserver::new(resize_closure.as_ref().unchecked_ref()).unwrap();

        // Clone the element ref and drop the borrow before calling observe(),
        // because observe() can fire the resize callback synchronously,
        // which would conflict with an active borrow.
        let el_clone = self.inner.borrow().el.clone();
        observer.observe(&el_clone);

        {
            let mut s = self.inner.borrow_mut();
            s.resize_observer = Some(observer);
            s.resize_closure = Some(resize_closure);
        }

        // Start with a pause then begin scrolling
        schedule_pause(&self.inner);
    }

    pub fn destroy(&self) {
        let mut s = self.inner.borrow_mut();
        s.stop_loop();

        if let Some(observer) = s.resize_observer.take() {
            observer.disconnect();
        }

        if let Some(ref closure) = s.mouseenter_closure {
            s.el
                .remove_event_listener_with_callback(
                    "mouseenter",
                    closure.as_ref().unchecked_ref(),
                )
                .ok();
        }
        if let Some(ref closure) = s.mouseleave_closure {
            s.el
                .remove_event_listener_with_callback(
                    "mouseleave",
                    closure.as_ref().unchecked_ref(),
                )
                .ok();
        }

        s.mouseenter_closure = None;
        s.mouseleave_closure = None;
        s.resize_closure = None;
        s.tick_closure = None;
        s.start_after_pause_closure = None;
    }
}
