function filterMovies() {
  // Declare variables
  var input, filters, filter, table, tr, a, i, j, txtValue;
  input = document.getElementsByClassName('searchTB')[0];
  filters = input.value.split(" ");
  table = document.getElementsByClassName("sortable")[0];
  tr = table.getElementsByTagName('tr');

  if (input.value == "") {
    for (i = 1; i < tr.length; i++) {
      tr[i].style.display = "";
    }
  }
  // Loop through all list items, and hide those who don't match the search query
  for (i = 1; i < tr.length; i++) {
    for (j = 0; j < filters.length; j++) {
      filter = filters[j].toUpperCase();
      a = tr[i].getElementsByTagName("td")[1];
      txtValue = a.textContent || a.innerText;
      if (txtValue.toUpperCase().indexOf(filter) > -1) {
        tr[i].style.display = "";
      } else {
        tr[i].style.display = "none";
        break
      }
    }
  }
}