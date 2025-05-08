// https://github.com/Smithay/input.rs

use input::event::keyboard::{KeyState, KeyboardEventTrait, KeyboardKeyEvent};
use input::event::KeyboardEvent;
use input::{Libinput, LibinputInterface};
use libc::{O_RDONLY, O_RDWR, O_WRONLY};
use std::fs::{File, OpenOptions};
use std::os::unix::{fs::OpenOptionsExt, io::OwnedFd};
use std::path::Path;

// Randomly copied boilerplate
struct Interface;

impl LibinputInterface for Interface {
    fn open_restricted(&mut self, path: &Path, flags: i32) -> Result<OwnedFd, i32> {
        OpenOptions::new()
            .custom_flags(flags)
            .read((flags & O_RDONLY != 0) | (flags & O_RDWR != 0))
            .write((flags & O_WRONLY != 0) | (flags & O_RDWR != 0))
            .open(path)
            .map(|file| file.into())
            .map_err(|err| err.raw_os_error().unwrap())
    }
    fn close_restricted(&mut self, fd: OwnedFd) {
        drop(File::from(fd));
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut input = Libinput::new_with_udev(Interface);
    input
        .udev_assign_seat("seat0")
        .expect("Couldn't assign seat");

    println!("code,time");
    loop {
        input.dispatch()?;
        for event in &mut input {
            match event {
                // 57 is space
                input::Event::Keyboard(KeyboardEvent::Key(ref ev))
                    if ev.key() == 57 && ev.key_state() == KeyState::Pressed =>
                {
                    // Subtle input lag
                    // requires subtraction
                    println!("{},{}", ev.key(), ev.time() - 2);
                }
                _ => {}
            }
        }
    }
}
