document.addEventListener('DOMContentLoaded', function() {
    console.log('script.js loaded');
    let searchForm = document.getElementById('search');
    console.log('searchForm:', searchForm);
    let pageLink = document.querySelectorAll('.page-link');
    console.log('pageLink count:', pageLink.length);


    if (searchForm) {
        for (let i = 0; pageLink.length > i; i++) {
            console.log('Attaching event listener to pageLink:', pageLink[i]);
            pageLink[i].addEventListener('click', function(e) {
                e.preventDefault()
                let page = this.dataset.page;

                searchForm.innerHTML += `<input value=${page} name="page" type="hidden">`

                searchForm.submit()
            })
        }
    }
});

