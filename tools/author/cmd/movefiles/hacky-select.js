window.hackySelect = [];
document.body.addEventListener('click', (e) => {
    e.preventDefault();

    let target = null;
    let targetTest = e.target;

    while (targetTest) {
        if (targetTest.nodeName == 'A') {
            target = targetTest;

            break;
        }

        targetTest = targetTest.parentElement;
    }

    if (!target) {
        return;
    }

    const hackyValue = (new URL(target.href)).pathname;
    const found = window.hackySelect.indexOf(hackyValue);
    if (found == -1) {
        window.hackySelect.push(hackyValue);
        target.style.opacity = 0.25;
    } else {
        window.hackySelect = window.hackySelect.filter(v => v != hackyValue);
        target.style.opacity = undefined;
    }
})