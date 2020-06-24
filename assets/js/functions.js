function areFiltersFound(trow, filters)
{
  tds = trow.getElementsByTagName("td");

  td = tds[1];
  txtValue = td.textContent || td.innerText;
  txtValue = txtValue.toUpperCase();

  for (j = 0; j < filters.length; j++) {
    filter = filters[j].toUpperCase();
    if (txtValue.indexOf(filter) < 0){
      // if we don't find any of the filters then return false
      return false;
    }
  }
  return true;
}
function filterMovies() {
  // Declare variables
  var input, filters, filter, table, tr, a, i, j, txtValue, nVis;
  nVis = 0;
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
  nVis = tr.length-1;
  for (i = 1; i < tr.length; i++) {
    if (areFiltersFound(tr[i], filters)) {
      tr[i].style.display = "";
    } else {
      tr[i].style.display = "none";
      nVis--;
    }
  }
  document.getElementById('numberOfMovies').innerHTML = nVis + ' Movies found';
}