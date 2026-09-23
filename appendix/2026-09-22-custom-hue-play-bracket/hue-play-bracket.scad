$fn = 256;

color("aliceblue") {
  translate([0, 0, 62]) {
    difference() {
      cube([40, 29.5, 30]);

      translate([0, 4, 0]) {
        cube([40, 26, 25.5]);
      }

      translate([0, 0, 30]) {
        rotate([-90, 0, 0]) {
          RoundedCorner();
        }
      }
    }
    
    translate([0, 29.5, 24]) {
      difference() {
        cube([40, 2.5, 6]);
      
        translate([0, 2.5, 6]) {
          rotate([180, 0, 0]) {
            RoundedCorner();
          }
        }
      }
    }
  }

  difference() {
    union() {
      cube([40, 25, 62]);

      cube([40, 54, 30]);
    }

    DoHueBarRotate() {
      HueBar();
      HueBarDrill();
    }
    
    RoundedCorner();
    
    translate([0, 54, 0]) {
      rotate([90, 0, 0]) {
        RoundedCorner();
      }
    }
  }
}

module RoundedCorner() {
  translate([-2.5, 1, 1]) {
    rotate([-90, 0, 0]) {
      rotate([0, 90, 0]) {
        difference() {
          cube([1, 1, 60]);
          linear_extrude(h=60) {
            circle(d=2);
          }
        }
      }
    }
  }
}

module DoHueBarRotate() {
  translate([0, 45, 5]) {
    rotate([55, 0, 0]) {
      translate([75, 35, 0]) {
        rotate([90, 0, 0]) {
          rotate([0, -90, 0]) {
            children();
          }
        }
      }
    }
  }
}

module HueBar() {
  color("dimgray") {
    linear_extrude(100) {
      polygon([[0, 0], [0, 46], [29, 44.5], [35, 23], [29, 1.5]]);
    }
  }
}

module HueBarDrill() {
  translate([0, 0, 55]) {
    union() {
      translate([34, 23, 0]) {
        rotate([0, 90, 0]) {
          linear_extrude(h=40) {
            circle(d=4);
          }
        }
      }

      translate([36, 23, 0]) {
        rotate([0, 90, 0]) {
          linear_extrude(h=40) {
            circle(d=6);
          }
        }
      }
    }
  }
}
