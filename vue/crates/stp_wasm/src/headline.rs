use std::cell::RefCell;
use std::rc::Rc;
use wasm_bindgen::prelude::*;
use web_sys::HtmlElement;

const SPEED: f64 = 0.5; // pixels per frame

struct Inner {
    container: HtmlElement,
    item1: HtmlElement,
    offset: f64,
    cached_width: f64,
    raf_id: Option<i32>,
    resize_observer: Option<web_sys::ResizeObserver>,
    // closures kept alive
    animate_closure: Option<Closure<dyn FnMut()>>,
    resize_closure: Option<Closure<dyn FnMut(js_sys::Array)>>,
}

#[wasm_bindgen]
pub struct HeadlineScroller {
    inner: Rc<RefCell<Inner>>,
}

impl Inner {
    fn measure_width(&mut self) {
        let container_width = self.container.offset_width() as f64;
        let item_width = self.item1.scroll_width() as f64;
        self.cached_width = container_width.max(item_width);
    }
}

fn schedule_frame(inner: &Rc<RefCell<Inner>>) {
    let inner_clone = Rc::clone(inner);

    let closure = Closure::once(move || {
        animate(&inner_clone);
    });

    let mut s = inner.borrow_mut();
    let window = web_sys::window().unwrap();
    let id = window
        .request_animation_frame(closure.as_ref().unchecked_ref())
        .unwrap();
    s.raf_id = Some(id);
    s.animate_closure = Some(closure);
}

fn animate(inner: &Rc<RefCell<Inner>>) {
    {
        let mut s = inner.borrow_mut();
        s.raf_id = None;

        if s.cached_width == 0.0 {
            drop(s);
            schedule_frame(inner);
            return;
        }

        s.offset -= SPEED;

        if s.offset <= -s.cached_width {
            s.offset += s.cached_width;
        }

        let transform = format!("translateX({}px)", s.offset);
        s.container.style().set_property("transform", &transform).ok();
    }

    schedule_frame(inner);
}

#[wasm_bindgen]
impl HeadlineScroller {
    #[wasm_bindgen(constructor)]
    pub fn new(container: HtmlElement, item1: HtmlElement) -> HeadlineScroller {
        let inner = Rc::new(RefCell::new(Inner {
            container,
            item1,
            offset: 0.0,
            cached_width: 0.0,
            raf_id: None,
            resize_observer: None,
            animate_closure: None,
            resize_closure: None,
        }));

        HeadlineScroller { inner }
    }

    pub fn start(&self) {
        // Measure initial width
        self.inner.borrow_mut().measure_width();

        // Set up resize observer
        let inner_clone = Rc::clone(&self.inner);
        let resize_closure = Closure::wrap(Box::new(move |_entries: js_sys::Array| {
            inner_clone.borrow_mut().measure_width();
        }) as Box<dyn FnMut(js_sys::Array)>);

        let observer =
            web_sys::ResizeObserver::new(resize_closure.as_ref().unchecked_ref()).unwrap();

        // Clone the element ref and drop the borrow before calling observe(),
        // because observe() can fire the resize callback synchronously,
        // which would conflict with an active borrow.
        let container_clone = self.inner.borrow().container.clone();
        observer.observe(&container_clone);

        {
            let mut s = self.inner.borrow_mut();
            s.resize_observer = Some(observer);
            s.resize_closure = Some(resize_closure);
        }

        // Start animation loop
        schedule_frame(&self.inner);
    }

    pub fn destroy(&self) {
        let mut s = self.inner.borrow_mut();

        let window = web_sys::window().unwrap();
        if let Some(id) = s.raf_id.take() {
            window.cancel_animation_frame(id).ok();
        }

        if let Some(observer) = s.resize_observer.take() {
            observer.disconnect();
        }

        s.animate_closure = None;
        s.resize_closure = None;
    }
}
