function IncomingCtrl($scope, $http, $location) {
  $http.get('/scan/incoming').success(function(data) {
    if (data.length > 0 && data[0].ID !== undefined) {
      $location.path('/doc/'+data[0].ID+'/incoming');
      return;
    }
    $scope.message = "no incoming documents";
  });
}

