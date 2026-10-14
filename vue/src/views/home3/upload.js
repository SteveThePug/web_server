/**
 * Uploading a file for use in a box: stores it through the chat's upload
 * endpoint (public, unguessable URLs under /uploads/) and says what markup
 * shows it. Used by BoxEditor's "Insert upload" and by dropping files onto
 * the canvas in edit mode.
 *
 * Never the private upload directory: that one is behind a cookie check, and
 * a box's sandboxed frame loads as an opaque origin, so its requests may not
 * carry the cookie.
 */
import { api } from "@/api";

/** Upload one file; resolves to its URL. Throws an ApiError on failure. */
export async function uploadFile(file) {
  const body = new FormData();
  body.append("file", file);
  const { url } = await api.post("/messages/upload", body, { fallback: `Upload of ${file.name} failed` });
  return url;
}

export const isImage = (url) => /\.(jpe?g|png|gif|webp)$/i.test(url);
export const isVideo = (url) => /\.(mp4|webm)$/i.test(url);

/** The markup an uploaded file is shown with, by type. Anything else (a PDF,
 *  a text file) becomes a link. */
export function tagFor(url) {
  if (isImage(url)) return `<img src="${url}" alt="">`;
  if (isVideo(url)) return `<video src="${url}" controls loop muted></video>`;
  if (/\.(mp3|ogg)$/i.test(url)) return `<audio src="${url}" controls></audio>`;
  return `<a href="${url}" target="_blank">${url.split("/").pop()}</a>`;
}

/** Natural size of an image or video file, or null if it can't be read. */
export function mediaSize(file) {
  return new Promise((resolve) => {
    const src = URL.createObjectURL(file);
    const done = (size) => {
      URL.revokeObjectURL(src);
      resolve(size);
    };
    if (file.type.startsWith("video/")) {
      const v = document.createElement("video");
      v.onloadedmetadata = () => done({ w: v.videoWidth, h: v.videoHeight });
      v.onerror = () => done(null);
      v.src = src;
    } else {
      const img = new Image();
      img.onload = () => done({ w: img.naturalWidth, h: img.naturalHeight });
      img.onerror = () => done(null);
      img.src = src;
    }
  });
}
