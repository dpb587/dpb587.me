$fn = $preview ? 32 : 512;

module rounded_cube(size, radius, height) {
  hull() {
    translate([radius, radius, 0]) {
      cylinder(h = height, r = radius);
    }

    translate([size[0] - radius, radius, 0]) {
      cylinder(h = height, r = radius);
    }

    translate([radius, size[1] - radius, 0]) {
      cylinder(h = height, r = radius);
    }

    translate([size[0] - radius, size[1] - radius, 0]) {
      cylinder(h = height, r = radius);
    }
  }
}

module rounded_cylinder(w, r, h) {
  hull() {
    translate([r, r, 0]) {
      cylinder(h = h, r = r);
    }

    translate([w - r, r, 0]) {
      cylinder(h = h, r = r);
    }
  }
}

difference() {
  rounded_cube([64, 64], 5, 2);

  // label
  translate([4, 4, 1.8]) {
    rounded_cube([56, 27], 2.5, 0.2);
  }

  // notch
  translate([17, 57, 0]) {
    rounded_cylinder(30, 2.5, 3);
  }

  translate([16.6, 56.6, 0]) {
    rounded_cylinder(30.8, 2.9, 0.2);
  }

  translate([16.8, 56.8, 0]) {
    rounded_cylinder(30.4, 2.7, 0.4);
  }

  translate([16.6, 56.6, 1.8]) {
    rounded_cylinder(30.8, 2.9, 0.2);
  }

  translate([16.8, 56.8, 1.6]) {
    rounded_cylinder(30.4, 2.7, 0.4);
  }
    
  // swatch
  translate([7, 38, 0.2]) cube([10, 6, 1.8]); 
  translate([12, 44, 0.4]) cube([10, 6, 1.6]);
  translate([17, 38, 0.6]) cube([10, 6, 1.4]);
  translate([22, 44, 0.8]) cube([10, 6, 1.2]);
  translate([27, 38, 1.0]) cube([10, 6, 1.0]);
  translate([32, 44, 1.2]) cube([10, 6, 0.8]);
  translate([37, 38, 1.4]) cube([10, 6, 0.6]);
  translate([42, 44, 1.6]) cube([10, 6, 0.4]);
  translate([47, 38, 1.8]) cube([10, 6, 0.2]);
}
