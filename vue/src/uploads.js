/**
 * Uploaded files: the one module to use whenever the frontend stores a file
 * or shows one it stored.
 *
 *   import { uploadFile, ACCEPT, kindOf, tagFor } from "@/uploads";
 *
 *   <input type="file" :accept="ACCEPT.images" @change="…">
 *   const url = await uploadFile(file);           // "/uploads/3f9c….png"
 *   kindOf(url)                                   // "image"
 *
 * Files go to POST /api/uploads (backend/handlers/handle_upload.go). The
 * backend decides everything that matters: the allowed types and the 10MB
 * limit (backend/services/uploads.go, UploadTypes), the random file name, and
 * where it is stored (the Garage bucket, falling back to disk). The URL it
 * returns is permanent and public — anyone with it can load the file, no login
 * — so never upload something that shouldn't be public unless it is a private
 * chat attachment (`{ private: true }`, admin only).
 *
 * A refused file (wrong type, too large) rejects with an ApiError whose
 * `message` is the backend's sentence, ready to show the user.
 */
import { api } from "@/api";

/**
 * `accept` values for file pickers, matching what the backend allows. They
 * only steer the picker; the backend enforces the real list.
 */
export const ACCEPT = {
  images: "image/jpeg,image/png,image/gif,image/webp",
  media: "image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm,audio/mpeg,audio/ogg",
  any: "image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm,audio/mpeg,audio/ogg,application/pdf,text/plain",
};

/**
 * Store one file; resolves to its URL.
 *
 * @param {File} file
 * @param {object} [options]
 * @param {boolean} [options.private]  admin-only: store it under
 *   /uploads/private/, readable only by admins. Never use it for anything
 *   shown inside a canvas box — a box's sandboxed frame loads without the
 *   login cookie, so the file would not load.
 * @returns {Promise<string>} "/uploads/<name>" or "/uploads/private/<name>"
 */
export async function uploadFile(file, { private: isPrivate = false } = {}) {
  const body = new FormData();
  body.append("file", file);
  if (isPrivate) body.append("private", "true");
  const { url } = await api.post("/uploads", body, { fallback: `Upload of ${file.name} failed` });
  return url;
}

/** What a stored file is, from its URL: "image", "video", "audio" or "file". */
export function kindOf(url) {
  if (/\.(jpe?g|png|gif|webp)$/i.test(url)) return "image";
  if (/\.(mp4|webm)$/i.test(url)) return "video";
  if (/\.(mp3|ogg)$/i.test(url)) return "audio";
  return "file";
}

export const isImage = (url) => kindOf(url) === "image";
export const isVideo = (url) => kindOf(url) === "video";

/**
 * Whether a URL is one of our own uploads. Use it before rendering a URL that
 * came from somebody else (a chat message's fileUrl arrives over the socket
 * and is otherwise attacker-controlled). A strict prefix check, so absolute
 * ("https://evil/uploads/x") and protocol-relative ("//evil/uploads/x") URLs
 * fail, and ".." is refused outright.
 */
export function isUploadUrl(url) {
  return typeof url === "string" && !url.includes("..") && url.startsWith("/uploads/");
}

/** The markup a stored file is shown with in a box's HTML, by kind. */
export function tagFor(url) {
  switch (kindOf(url)) {
    case "image":
      return `<img src="${url}" alt="">`;
    case "video":
      return `<video src="${url}" controls loop muted></video>`;
    case "audio":
      return `<audio src="${url}" controls></audio>`;
    default:
      return `<a href="${url}" target="_blank">${url.split("/").pop()}</a>`;
  }
}

/** Natural size of an image or video File, or null if it can't be read. */
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
