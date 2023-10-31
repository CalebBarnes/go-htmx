const autoprefixer = require("autoprefixer");
const postcss = require("postcss");
const postcssNested = require("postcss-nested");
const fs = require("fs");

fs.readFile("src/main.css", (err, css) => {
  postcss([autoprefixer, postcssNested])
    .process(css, { from: "src/main.css", to: "static/css/main.css" })
    .then((result) => {
      fs.writeFile("static/css/main.css", result.css, () => true);
      if (result.map) {
        fs.writeFile(
          "static/css/main.css.map",
          result.map.toString(),
          () => true
        );
      }
    });
});
