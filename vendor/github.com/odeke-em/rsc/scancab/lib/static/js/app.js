'use strict';

/* App Module */

angular.module('scancab', []).
  config(['$routeProvider', function($routeProvider) {
  $routeProvider.
      when('/', {templateUrl: 'partials/list.html', controller: ListCtrl}).
      when('/incoming', {templateUrl: 'partials/incoming.html', controller: IncomingCtrl}).
      when('/doc/:docid', {templateUrl: 'partials/show.html', controller: ShowCtrl}).
      when('/doc/:docid/edit', {templateUrl: 'partials/edit.html', controller: EditCtrl}).
      when('/doc/:docid/incoming', {templateUrl: 'partials/edit.html', controller: EditIncomingCtrl}).
      otherwise({redirectTo: '/'});
}]);
