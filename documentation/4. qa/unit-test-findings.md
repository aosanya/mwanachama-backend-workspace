# Unit test findings
- manager_cover_test.go (TestGetWorkbook_AnswersNotFoundForADeletedWorkbook): GetWorkbook returns a workbook after DeleteWorkbook marked it deleted; expected ErrWorkbookNotFound, so delete, update, record and touch on it are refused too.
