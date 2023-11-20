import htmx from "htmx.org";

declare global {
  interface Window {
    htmx: typeof htmx;
  }
}

window.htmx = htmx;

// import("htmx.org/dist/ext/preload")
//   .then(() => {
//     // Code from the imported script can be executed here
//   })
//   .catch((error) => {
//     console.error("Failed to load the script:", error);
//   });

// import("htmx.org/dist/ext/head-support")
//   .then(() => {
//     // Code from the imported script can be executed here
//   })
//   .catch((error) => {
//     console.error("Failed to load the script:", error);
//   });
