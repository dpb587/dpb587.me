const autoprefixer = require('autoprefixer')
const gulp = require('gulp');
const postcss = require('gulp-postcss')
const purgecss = require('gulp-purgecss')
const sourcemaps = require('gulp-sourcemaps')
const tailwindcss = require('tailwindcss')

gulp.task('css:dev', function () {
  return gulp.src('./assets/src/css/*.css')
    .pipe(sourcemaps.init())
    .pipe(postcss([
      tailwindcss,
      autoprefixer(),
    ]))
    .pipe(sourcemaps.write('.'))
    .pipe(gulp.dest('./assets/dist/css'));
});

gulp.task('css:prod', function () {
  return gulp.src('./assets/src/css/*.css')
    .pipe(postcss([
      tailwindcss,
      autoprefixer(),
    ]))
    .pipe(purgecss({
      content: [
        './layouts/**/*.html',
      ],
      defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || [],
    }))
    .pipe(gulp.dest('./assets/dist/css'));
});
