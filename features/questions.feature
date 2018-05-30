Feature: Knowledge Base question retrieval
  Questions should be retrievable through our api

Scenario: The application has a get questions endpoint
  Given I do have a running web server
  And Questions already existin the db
  When I request to retrieve questions
  Then I should see a JSON array response

