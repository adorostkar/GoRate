function areFiltersFound(trow, filters)
{
  var tds, td, i, j, filter, txtValue, searchColumns;
  // [0, 1] ==> title, genre
  // [0] ==> title
  // [1] ==> genre
  searchColumns = [1]; 
  tds = trow.getElementsByTagName("td");


  for (i = 0; i < filters.length; i++) {
    filter = filters[i].toUpperCase();
    for (j = 0; j < searchColumns.length; j++) { 
      td = tds[searchColumns[j]];
      txtValue = td.textContent || td.innerText;
      txtValue = txtValue.toUpperCase();
      if (txtValue.indexOf(filter) > -1){
        break;
      } else {
        // if we don't find any of the filters then return false
        return false;
      }
    }
  }
  return true;
}
function filterMovies() {
  // Declare variables
  var input, filters, table, tr, i, nVis;
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