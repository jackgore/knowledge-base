Feature: Knowledge Base Team Creation
  In order to gather knowledge
  The users
  Should be able to create organizations and teams

Scenario Outline: The application has a create organizations endpoint
  Given I do have a running web server
  When I post a <isValid> organization
  Then I should see a <code> response

  Examples:
	| isValid | code |
	| true    | 200  |
	| false   | 400  |
	
Scenario Outline: The application has a get organization endpoint
  Given I do have a running web server
  When I try to retrieve organization with name <name>
  Then I should see a <code> response

  Examples:
	| name    | code |
	| fake    | 404  |
	| real    | 200  |
	
Scenario Outline: The application has a create team endpoint
  Given I do have a running web server
  When I post a <orgValid> organization
  And I post a <teamValid> team
  Then I should see a <code> response

  Examples:
	| orgValid | teamValid | code |
	| true     | true      | 200  |
	| true     | false     | 400  |
	| false    | false     | 400  |
	| false    | true      | 400  |
	
	
