package script

var ScriptMain map[string]string

func GetFilterExceptions() string {
	return `
	SET NOCOUNT ON
	
	DELETE FROM Analiz_EN.dbo.FirmaReadyABM WHERE FirmaID IN (99,96,71)

	if object_id('tempdb..#tmp') is not null 
		  DROP TABLE #tmp 
	   
	create table #tmp (val char(9), isfolder tinyint, primary key clustered (val))

	INSERT INTO #tmp(val, isfolder) 
	SELECT 
		t.ID, t.ISFOLDER 
	FROM $Справочник.Номенклатура as t 
	WHERE CAST(t.CODE as int) IN (27450,9545,93307,324129,281256,107388, 93585, 93206, 67197, 74,318970,318971,319311,93089,824455,92599)

	exec dbo.PutObjectListTovar '#tmp'

	if object_id('tempdb..#ТоварыИсключения') is not null   
		  DROP TABLE #ТоварыИсключения   
   
--	CREATE TABLE #ТоварыИсключения (ArticleID int, primary key clustered (ArticleID))
	CREATE TABLE #ТоварыИсключения (val char(9), isfolder tinyint, primary key clustered (val))

--	INSERT INTO #ТоварыИсключения (ArticleID) 
	INSERT INTO #ТоварыИсключения (val, isfolder) 
	SELECT 
--	   CAST(t.CODE as int) as ArticleID 
	   t.ID, t.ISFOLDER
	FROM #tmp 
	INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON t.ID=#tmp.Val
	WHERE t.ISFOLDER=2

	exec dbo.PutObjectListTovar '#ТоварыИсключения'

	SET NOCOUNT OFF
	`
}

func GetSnapManagers() string {
	return `
	SET NOCOUNT ON

	IF (NOT (OBJECT_ID('tempdb..#ArticleManagersMain') IS NULL))
		DROP TABLE #ArticleManagersMain
	
	SELECT  DISTINCT
	  t.ID as Группа
--	 ,m.Manager
	 ,am.Manager
	 ,am.Сотрудник
--	 ,ISNULL(s.ID,'     0   ') as Сотрудник
	INTO #ArticleManagersMain
--	FROM Analiz_EN.dbo.ArticleManagersMain am
	FROM ArticleManagerLink am (NOLOCK)
	INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON t.ID=am.Группа -- CAST(t.CODE as int)=am.ArticleID
--	INNER JOIN Analiz_EN.dbo.ManagersMain m ON m.ManagerID=am.ManagerID
--	LEFT OUTER JOIN $Справочник.Сотрудники as s (NOLOCK) ON LTRIM(RTRIM(s.DESCR))=LTRIM(RTRIM(m.Manager))

	CREATE INDEX __X__ArticleManagersMain_GROUP_1 ON #ArticleManagersMain(Группа)

	DECLARE @rw int, @i int

	SET @rw=1
	SET @i=0

	WHILE (@rw <> 0 AND @i < 18)
	BEGIN
		INSERT INTO #ArticleManagersMain (Группа, Manager, Сотрудник)
		SELECT  DISTINCT t.ID, M.Manager, M.Сотрудник
		FROM #ArticleManagersMain AS M
		INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON t.ParentID=M.Группа
		LEFT JOIN #ArticleManagersMain AS P ON P.Группа = t.ID
		WHERE P.Группа IS NULL AND t.ISFOLDER=1
	
	    SET @rw=@@ROWCOUNT
	    SET @i=@i+1

	END

	IF (NOT (OBJECT_ID('_ArticleManagersMain') IS NULL))
		DROP TABLE _ArticleManagersMain

	SELECT * INTO _ArticleManagersMain FROM #ArticleManagersMain

	SET NOCOUNT OFF
	`
}

func InitScript() {
	ScriptMain = make(map[string]string)
	ScriptMain["ФильтрИсключений"] = GetFilterExceptions()
	ScriptMain["ПривязкуМенеджеров"] = GetSnapManagers()
}
