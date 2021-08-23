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

func GetListOfProductsFromLaunchedGroups() string {
	return `

	INSERT INTO #ТоварыПодключения (val, isfolder)
	SELECT
		M.Category, t.ISFOLDER
	FROM ConnectedCategories M (NOLOCK)
	INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON t.ID=M.Category

	exec dbo.PutObjectListTovar '#ТоварыПодключения'

	INSERT INTO #ConnectedCategories (Товар, Сотрудник)
	SELECT
		 t.ID Товар 
		,ISNULL(am.Сотрудник,'     0   ') as Сотрудник
	FROM #ТоварыПодключения tp
	INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON t.ID=tp.Val
	LEFT OUTER JOIN #ArticleManagersMain as am ON am.Группа=t.ParentID
	WHERE CAST(t.CODE as int) NOT IN (808037,808043,808675,808676,808677,808678)

	CREATE INDEX _IX_ConnectedCategories_ID ON #ConnectedCategories(Товар)

	`
}

func GetNumberOfStoresDownloaded() string {
	return `
	SELECT 
		COUNT(*) c 
	FROM Analiz_EN.dbo.FirmaReadyABM fr (NOLOCK) 
	INNER JOIN Analiz_EN.dbo.FirmaEN as f ON f.FirmaID=fr.FirmaID
	LEFT OUTER JOIN  
	( 
			SELECT 
				 af.FirmaID 
				,COUNT(*) c 
			FROM Analiz_EN.dbo.ArticleFromShopForAMBDaily af (NOLOCK)
			INNER JOIN Analiz_EN.dbo.FirmaEN as f ON f.FirmaID=af.FirmaID 
			INNER JOIN $Справочник.Номенклатура as a (NOLOCK) ON CAST(a.CODE as int)=af.SKU 
			WHERE af.Data = :Дат AND a.ISFOLDER=2 AND a.ISMARK=0  AND f.Dostup=2
			GROUP BY af.FirmaID 
	) X ON X.FirmaID=fr.FirmaID  
	WHERE X.FirmaID IS NOT NULL  AND f.Dostup=2
	`
}

func GetNumberOfNotArrivedStores() string {
	return `
	SELECT 
		COUNT(*) c 
	FROM Analiz_EN.dbo.FirmaReadyABM fr (NOLOCK)
	INNER JOIN Analiz_EN.dbo.FirmaEN as f ON f.FirmaID=fr.FirmaID 
	LEFT OUTER JOIN  
	( 
			SELECT 
				 af.FirmaID 
				,COUNT(*) c 
			FROM Analiz_EN.dbo.ArticleFromShopForAMBDaily af (NOLOCK)
			INNER JOIN Analiz_EN.dbo.FirmaEN as f ON f.FirmaID=af.FirmaID  
			INNER JOIN $Справочник.Номенклатура as a (NOLOCK) ON CAST(a.CODE as int)=af.SKU 
			WHERE af.Data = :Дат AND a.ISFOLDER=2 AND a.ISMARK=0  AND f.Dostup=2
			GROUP BY af.FirmaID 
	) X ON X.FirmaID=fr.FirmaID  
	WHERE X.FirmaID IS NULL AND f.Dostup=2 
	`
}

func GetGetListOfNotArrivedStores() string {
	return `
	SELECT 
	$f.КраткоеНазвание as КраткоеНазвание 
FROM Analiz_EN.dbo.FirmaReadyABM fr (NOLOCK) 
INNER JOIN TD_EN.dbo.$Справочник.Фирмы as f (NOLOCK) ON CAST(f.CODE as int)=fr.FirmaID 
INNER JOIN TD_EN.dbo.$Справочник.Супермаркеты as sp (NOLOCK) ON sp.ID=$f.Супермаркет
LEFT OUTER JOIN  
( 
	SELECT 
		 af.FirmaID 
		,COUNT(*) c 
	FROM Analiz_EN.dbo.ArticleFromShopForAMBDaily af  (NOLOCK)
INNER JOIN Analiz_EN.dbo.FirmaEN as f ON f.FirmaID=af.FirmaID  
	INNER JOIN TD_EN.dbo.$Справочник.Номенклатура as a (NOLOCK) ON CAST(a.CODE as int)=af.SKU 
	WHERE af.Data >= :Дат AND af.Data < DATEADD(d,1,:Дат) AND f.Dostup=2 
	GROUP BY af.FirmaID 
) X ON X.FirmaID=fr.FirmaID 
WHERE X.FirmaID IS NULL  AND $sp.Доступен=2
ORDER BY $f.КраткоеНазвание
`
}

func GetGeneratingSampleData() string {
	return `
	SET NOCOUNT ON

	DECLARE @datDefault varchar(10), @dt1900 DateTime, @dt1754 DateTime, @DataAds DateTime
	
	SET @DataAds=(SELECT TOP 1 Data FROM Analiz_EN.dbo.AdsHalfYear (NOLOCK) ORDER BY Data DESC)
	
	SET @datDefault=CONVERT(varchar(10),DATEADD(d,-1,GETDATE()),102)
	SET @dt1900=CAST('19000101' as DateTime)
	SET @dt1754=CAST('17540101' as DateTime)
	
	SET NOCOUNT OFF
	
	SELECT
	
		 CASE WHEN am.Data is NULL THEN @datDefault ELSE CONVERT(varchar(10),am.Data,102) END as Data
		,ls.ArticleID SKU
		,ls.FirmaID
		,ISNULL(	
			CASE WHEN ISNULL(aff.ClientID,0) = 0 
				THEN 0 
				ELSE  
	--		    	    CASE WHEN sb.Tovar is NULL THEN 0 ELSE CASE WHEN sb.SB < 0 THEN 0 ELSE am.Active END END
						CASE WHEN sb.Tovar is NULL THEN 0 ELSE CASE WHEN sb.SB < 0 THEN 0 ELSE CASE WHEN am.Active is NULL THEN ls.Active ELSE am.Active END END END
			 END 
		 ,0 ) as Active
		   ,CASE WHEN am.Active is NULL THEN ls.Active ELSE am.Active END as WhAdditional_1
		,ISNULL(amp.ERP,0) as WhAdditional_2
		,CASE WHEN Ads.Quantity is NULL THEN 0.0 ELSE Ads.Quantity END as WhAdditional_3
		,CASE WHEN am.DataLastInput=@dt1900 OR am.DataLastInput=@dt1754 THEN '' ELSE CONVERT(varchar(10),am.DataLastInput,102) END as WhAdditional_4
		, ISNULL(aff.ClientID,0) as ClientID
		, 1 as UseClient
		,ISNULL(am.ClientID,0) as ClientLastPrihod
		,CASE WHEN ISNULL(sb.MinZakaz,0)=0 THEN ISNULL(am.MOQ,0) ELSE sb.MinZakaz END as MOQ
		,$t.КвантЗаказа as USQ
		,ISNULL(am.PurchasePrice,0) PurchasePrice
		,ISNULL(am.SalePrice,0) SalePrice
		,am.Code
		,'' as [Description]
		,ISNULL(gm.OrderSplitMark,'') as OrderSplitMark
		,ISNULL($sp.ИмяВКонфигураторе,'') as UserCode					-- Код сотрудника
		,'' as Buffer													-- Первоначальный буфер безопасности
		,CASE WHEN ISNULL(aff.ClientID,0) = 0 
			THEN 0.0 
			ELSE  
				CASE WHEN sb.Tovar is NULL THEN 0.0 ELSE CASE WHEN sb.SB < 0 THEN 0.0 ELSE sb.SB END END
		 END as SB_1													-- Витрина
		,'' as SB_2
		,'' as SB_3
		,'' as SB_4
		,'' as SB_5
		,CASE 
			WHEN am.Active = 0    THEN 'NA'
			WHEN $t.Ассортимент=1 THEN 'MTS'
			ELSE 'MTO'
		 END as [state]
	--	,CASE WHEN ls.FirmaID=99 THEN ISNULL(rc.Ostatok,0.000) ELSE ISNULL(am.Ostatok,0) END as Ostatok
		,ISNULL(am.Ostatok,0) as Ostatok
		,ISNULL(am.Prihod_Qnt,0) as Prihod_Qnt
		,ISNULL(am.Return_Qnt,0) as Return_Qnt
		,ISNULL(am.Sale_Qnt,0) as Sale_Qnt
		,ISNULL(am.Return_Qnt_In,0) as Return_Qnt_In
		,ISNULL(am.Qnt_Out_Move,0) as Qnt_Out_Move
		,0 as Qnt_Out_Manufacture
		,ISNULL(am.Qnt_Out_WriteOff,0) as Qnt_Out_WriteOff
		,ISNULL(am.ERP,0) as Season
		,$t.Пассивный as Passiv
		,$t.КвантЗаказа as КвантЗаказа
		,$t.Фреш as Fresh
		,CASE WHEN asn.ArticleID is NULL AND asn.FirmaID is NULL THEN 0 ELSE 1 END as Exclused
		,CASE WHEN afsn.ArticleID is NULL AND afsn.FirmaID is NULL THEN 0 ELSE 1 END as ExclusedActive
		,$f.КраткоеНазвание as NameShort
		,ISNULL(am.Sale_SumRozn,0.0) as Sale_SumRozn
		,ISNULL(am.Sale_SumSeb,0.0) as Sale_SumSeb
	INTO #__tmp__
	
	FROM #ListYesterdayBody ls
	INNER JOIN Analiz_EN.dbo.FirmaReadyABM  fra (NOLOCK) ON fra.FirmaID=ls.FirmaID
	INNER JOIN #ListSkuAll la ON la.ArticleID=ls.ArticleID
	INNER JOIN $Справочник.Фирмы as f (NOLOCK) ON CAST(f.CODE as int)=ls.FirmaID
	INNER JOIN $Справочник.Номенклатура as t (NOLOCK) 
		ON CAST(t.CODE as int)=ls.ArticleID
	INNER JOIN $Справочник.ЕдиницыИзмерений as e (NOLOCK) ON e.ID=$t.БазоваяЕдиница
	--LEFT OUTER JOIN #ОстаткиРЦ rc ON rc.ArticleID=ls.ArticleID
	LEFT OUTER JOIN Analiz_EN.dbo.ArticleFromShopForAMBDaily am (NOLOCK) ON am.SKU=ls.ArticleID AND am.Data=:Дат AND am.FirmaID=ls.FirmaID
	LEFT OUTER JOIN Analiz_EN.dbo.ArticleFromShopForAMBDaily amp (NOLOCK) ON amp.SKU=ls.ArticleID AND amp.FirmaID=ls.FirmaID
			AND  amp.Data = :ДатПред
	LEFT OUTER JOIN #ArticleManagersMain as amm ON amm.Группа=t.ParentID
	LEFT OUTER JOIN $Справочник.Сотрудники as sp (NOLOCK) ON sp.ID=ISNULL(amm.Сотрудник,$ПустойИД) AND ISNULL($sp.ИмяВКонфигураторе,'') <> ''
	LEFT OUTER JOIN SecurityBuffer as sb (NOLOCK) ON sb.Tovar=t.ID AND sb.Shop=f.ID
	LEFT OUTER JOIN Analiz_EN.dbo.TovarShopClient as aff (NOLOCK) ON aff.ArticleID=ls.ArticleID AND aff.FirmaID=ls.FirmaID
	LEFT OUTER JOIN
		(
			SELECT
				 CAST(kk01.CODE as int) ClientID
				,CAST(ISNULL(kk02.CODE,'0') as int) MainDepID
			FROM $Справочник.Контрагенты kk01 (NOLOCK)
			LEFT OUTER JOIN $Справочник.Контрагенты kk02 (NOLOCK) ON kk02.ID=$kk01.ГоловноеПредприятие
		) kt ON kt.ClientID=aff.ClientID
	LEFT OUTER JOIN Analiz_EN.dbo.AdsHalfYear Ads (NOLOCK) ON Ads.ArticleID=ls.ArticleID AND Ads.FirmaID=ls.FirmaID AND Ads.Data=@DataAds
	LEFT OUTER JOIN Analiz_EN.dbo.TovarShopClientZakaz as gm (NOLOCK) 
				ON gm.ClientID=ISNULL(aff.ClientID,0) AND gm.FirmaID=ISNULL(aff.FirmaID,0) AND gm.ArticleID=ISNULL(aff.ArticleID,0)
	LEFT OUTER JOIN Analiz_EN.dbo.ArticleShopNotUpload asn (NOLOCK) ON asn.ArticleID=ls.ArticleID AND asn.FirmaID=ls.FirmaID
	LEFT OUTER JOIN Analiz_EN.dbo.ArticleShopActiveNotUpload afsn (NOLOCK) ON afsn.ArticleID=ls.ArticleID AND afsn.FirmaID=ls.FirmaID
	--LEFT OUTER JOIN Analiz_EN.dbo.ArticleForABMNotUpload anu (NOLOCK) ON  anu.ArticleID=ls.ArticleID AND anu.FirmaID=ls.FirmaID AND  ISNULL(sb.SB,0.0)=0.0 
	LEFT OUTER JOIN Analiz_EN.dbo.ArticleForABMNotUpload anu (NOLOCK) ON  anu.ArticleID=CAST(t.CODE as int) AND anu.FirmaID=CAST(f.CODE as int) AND  ISNULL(sb.SB,0.0)=0.0 
	WHERE  --ls.FirmaID = :FirmaID AND
				  t.ID NOT IN (SELECT val FROM #ТоварыИсключения) 
			  AND t.ISMARK=0 AND t.ISFOLDER=2
			  AND $t.Уцененный=0
			  AND $t.Фреш=0
			  AND 
				 ( anu.ArticleID is NULL  AND anu.FirmaID   is NULL -- AND  ISNULL(sb.SB,0.0) > 0.0 
				 )
				  
		`
}

func GetStores() string {
	return `
SELECT
	CAST(fa.FirmaID as varchar)  FirmaID
   ,RTRIM(f.DESCR) Name
   ,ISNULL(fp.DESCR,'') group_name
   ,CASE WHEN ff.Dostup=2 THEN '1' ELSE '0' END as active
   ,CASE WHEN fa.FirmaID=99 THEN '1' ELSE '0' END as central
   ,fa.Addres	Addres
   ,CAST(ROUND(fa.geo_lat,7) as varchar)	latitude
   ,CAST(ROUND(fa.geo_lon,7) as varchar)	longitude
   ,'0'			deleted
   ,CONVERT(varchar(10),ff.DataStart,102)	open_at
   ,''			close_at
   ,fa.region	region
   ,'1'			in_shelf
FROM Analiz_EN.[dbo].[FirmaReadyABM] fa (NOLOCK)
INNER JOIN $Справочник.Фирмы f (NOLOCK) ON  fa.FirmaID=CAST(f.CODE as int)
INNER JOIN Analiz_EN.[dbo].[Firma] ff (NOLOCK) ON ff.FirmaID=fa.FirmaID
LEFT OUTER JOIN $Справочник.Фирмы fp (NOLOCK) ON fp.ID=f.ParentID
WHERE fa.FirmaID NOT IN (99,96,71,137) AND ff.Dostup=2
 AND CAST(ISNULL(fp.CODE,'0') as int) <> 168 -- ШОУ-РУМ
   `
}

func GetSkuHeader() string {
	return `
	SET NOCOUNT ON
	DECLARE @str varchar(300), @lev int

	SET @str=''
	SET @lev=0


	INSERT INTO #ArticleFilter (ArticleID)
	SELECT DISTINCT
		SKU
	FROM #__tmp__ ls

SET NOCOUNT OFF


	SELECT
		 SKU, Name
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,1)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,1) END as Additional_1
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,2)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,2) END as Additional_2
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,3)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,3) END as Additional_3
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,4)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,4) END as Additional_4
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,5)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,5) END as Additional_5
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,6)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,6) END as Additional_6
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,7)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,7) END as Additional_7
		,B.Additional_8
		,B.Additional_9
		,B.Additional_10
		,B.Additional_11
		,B.Additional_12
--		,ISNULL(rc.Ostatok,0.000) as Additional_13
		,0.000 as Additional_13
		,CASE WHEN cc.Товар is NULL THEN 0 ELSE 1 END as Additional_14
		,B.Dimension
		,B.Measure1
		,0  as Measure2
		,B.Passiv
		,B.Fresh
	FROM
	(
		SELECT
			 a.ArticleID as SKU
			,t.DESCR as Name
			,REPLACE(LTRIM(RTRIM(dbo.GetFullPath_$Справочник.Номенклатура(t.ID,'',0))),LTRIM(RTRIM(t.DESCR)),'') as FullPath
			,CASE WHEN $t.ТипТовара = $Перечисление.ТипыТоваров.Весовой THEN 'весовой' ELSE 'невесовой' END as Additional_8
			,ISNULL(LTRIM(RTRIM(s.DESCR)),'') as Additional_9
			,$t.СТМ as Additional_10
			,$t.VIP as Additional_11
			,$t.Ассортимент as Additional_12
			,LTRIM(RTRIM(e.DESCR)) as Dimension
			,$t.Пассивный as Passiv
			,CASE WHEN $t.ТипТовара = $Перечисление.ТипыТоваров.Весовой THEN 1 ELSE $t.КоэффициентВеса END as Measure1
			,$t.Фреш as Fresh
			,t.ID
		FROM #ArticleFilter as a
		INNER JOIN #ListSkuAll la ON la.ArticleID=a.ArticleID
		INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON CAST(t.CODE as int) =a.ArticleID
		INNER JOIN $Справочник.ЕдиницыИзмерений as e (NOLOCK) ON e.ID=$t.БазоваяЕдиница
		LEFT OUTER JOIN $Справочник.Страны as s (NOLOCK) ON s.ID=$t.Страна AND t.ISMARK=0 AND t.ISFOLDER=2
	) B
	LEFT OUTER JOIN #ConnectedCategories as cc ON cc.Товар=B.ID
--	LEFT OUTER JOIN #ОстаткиРЦ rc ON rc.ArticleID=B.SKU
	ORDER BY B.SKU
	`
}

func GetSkuHeaderNew() string {
	return `
	SET NOCOUNT ON
	DECLARE @str varchar(300), @lev int

	SET @str=''
	SET @lev=0

	TRUNCATE TABLE #ArticleFilter

	INSERT INTO #ArticleFilter (ArticleID)
	SELECT DISTINCT
		SKU
	FROM #__tmp__ ls

	if object_id('tempdb..#ArticleGroup') is not null
		  DROP TABLE #ArticleGroup
	CREATE TABLE #ArticleGroup (ArticleID int, primary key clustered (ArticleID))

	INSERT INTO #ArticleGroup (ArticleID)
	SELECT DISTINCT t.ParentCode
	FROM #ArticleFilter aa
	INNER JOIN Analiz_EN.dbo.Tovar t (NOLOCK) ON t.ArticleID=aa.ArticleID

	exec Analiz_EN.dbo.PutObjectListGroup '#ArticleGroup'

	INSERT INTO #ArticleFilter
	SELECT ArticleID FROM #ArticleGroup

SET NOCOUNT OFF


	SELECT
		 SKU, Name, ArticleGroupID
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,1)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,1) END as Additional_1
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,2)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,2) END as Additional_2
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,3)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,3) END as Additional_3
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,4)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,4) END as Additional_4
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,5)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,5) END as Additional_5
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,6)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,6) END as Additional_6
		,CASE WHEN LTRIM(RTRIM(dbo.GetGroupFromLevel(FullPath,7)))=LTRIM(RTRIM(Name)) THEN '' ELSE dbo.GetGroupFromLevel(FullPath,7) END as Additional_7
		,B.Additional_8
		,B.Additional_9
		,B.Additional_10
		,B.Additional_11
		,B.Additional_12
--		,ISNULL(rc.Ostatok,0.000) as Additional_13
		,0.000 as Additional_13
		,CASE WHEN cc.Товар is NULL THEN 0 ELSE 1 END as Additional_14
		,B.Dimension
		,B.Measure1
		,0  as Measure2
		,B.Passiv
		,B.Fresh
		,CAST(B.ISFOLDER as int) ISFOLDER, B.main_dimension_uid, B.Fractional
		,ISNULL(vn.property,'') brand_uid
	FROM
	(
		SELECT
			 a.ArticleID as SKU
			,ISNULL(CAST(pt.CODE as int),0) ArticleGroupID
			,t.DESCR as Name
			,REPLACE(LTRIM(RTRIM(dbo.GetFullPath_$Справочник.Номенклатура(t.ID,'',0))),LTRIM(RTRIM(t.DESCR)),'') as FullPath
			,CASE 
			   WHEN t.ISFOLDER = 1 THEN ''
			   ELSE
				CASE WHEN $t.ТипТовара = $Перечисление.ТипыТоваров.Весовой THEN 'весовой' ELSE 'невесовой' END
			 END  as Additional_8
			,ISNULL(LTRIM(RTRIM(s.DESCR)),'') as Additional_9
			,$t.СТМ as Additional_10
			,$t.VIP as Additional_11
			,$t.Ассортимент as Additional_12
			,LTRIM(RTRIM(ISNULL(e.DESCR,''))) as Dimension
			,$t.Пассивный as Passiv
			,CASE WHEN $t.ТипТовара = $Перечисление.ТипыТоваров.Весовой THEN 1 ELSE $t.КоэффициентВеса END as Measure1
			,$t.Фреш as Fresh
			,t.ID
			,t.ISFOLDER
--			,ISNULL(LTRIM(RTRIM(e.CODE)),'') as main_dimension_uid  	-- Уникальный идентификатор ОСНОВНОЙ единицы измерения
			,CASE WHEN e.CODE is NULL THEN '' ELSE LTRIM(RTRIM(t.CODE))+'/'+LTRIM(RTRIM(e.CODE)) END as main_dimension_uid  	-- Уникальный идентификатор ОСНОВНОЙ единицы измерения
			, CASE WHEN $t.ТипТовара=$Перечисление.ТипыТоваров.Штучный THEN 0 ELSE 1 END as Fractional
		FROM #ArticleFilter as a
		LEFT OUTER JOIN #ListSkuAll la ON la.ArticleID=a.ArticleID
		INNER JOIN $Справочник.Номенклатура as t (NOLOCK) ON CAST(t.CODE as int) =a.ArticleID
		LEFT OUTER JOIN $Справочник.Номенклатура as pt (NOLOCK) ON pt.ID=t.ParentID
		LEFT OUTER JOIN $Справочник.ЕдиницыИзмерений as e (NOLOCK) ON e.ID=$t.БазоваяЕдиница
		LEFT OUTER JOIN $Справочник.Страны as s (NOLOCK) ON s.ID=$t.Страна AND t.ISMARK=0 AND t.ISFOLDER=2
		WHERE t.ISFOLDER=1 OR la.ArticleID is NOT NULL
	) B
	LEFT OUTER JOIN #ConnectedCategories as cc ON cc.Товар=B.ID
	LEFT OUTER JOIN Analiz_EN.[dbo].[_IM_GetVendor] vn (NOLOCK) ON vn.ArticleID=B.SKU
--	LEFT OUTER JOIN #ОстаткиРЦ rc ON rc.ArticleID=B.SKU
	ORDER BY B.SKU
	`
}

func GetBrands() string {
	return `
SELECT
	*
FROM
(
	SELECT DISTINCT
		property, property_name
	FROM Analiz_EN.[dbo].[_IM_GetVendor] vn
	INNER JOIN #ArticleFilter a ON a.ArticleID=vn.ArticleID
) A
ORDER BY A.property_name
	`
}

func GetSuppliers() string {
	return `
SELECT 
	am.ClientID
   ,CASE WHEN $k.Пассивный=1 THEN 0 ELSE 1 END as Active
   ,LTRIM(RTRIM(k.DESCR)) as Name
   ,$ПоследнееЗначение.Контрагенты.ПочтовыйАдрес(k.ID,GETDATE()) as Address
   ,$k.Телефоны as Phone
   ,$k.EMail as EMail
FROM 
(
   SELECT DISTINCT XX.ClientID
   FROM
   (
	   SELECT ClientID FROM #ClientFilter

	   UNION ALL

	   SELECT ClientID FROM ClientLife

	   UNION ALL

	   SELECT
		   CAST(k.CODE as int) as ClientID
	   FROM $Справочник.ГрафикЗаказовАВМ as a (NOLOCK)
	   INNER JOIN $Справочник.Контрагенты as k (NOLOCK) ON k.ID=$a.Поставщик
	   WHERE a.ISMARK = 0
   ) XX
) am
INNER JOIN $Справочник.Контрагенты as k (NOLOCK) ON CAST(k.CODE as int)=am.ClientID
WHERE am.ClientID <> 0
ORDER BY k.DESCR
	`
}

func GetSchedule() string {
	return `
SELECT
	sz.CODE
   , f.Code as FirmaID, $f.КраткоеНазвание as fname
   , k.CODE as SupplierCode, k.DESCR as Поставщик
   , $sz.РазделительЗаказов as OrderSplitMark
   , CASE WHEN $sz.Активный=0 THEN 0 ELSE $sgm.Активный END as  Активный
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.Периодичность ELSE 7 END as Периодичность  
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Понедельник > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Monday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Вторник > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as TuesDay
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Среда > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Wednesday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Четверг > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Thursday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Пятница > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Friday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Суббота > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Saturday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Воскресенье > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Sunday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.ЧастотаФормирования ELSE CASE WHEN $sz.Периодичность=7 THEN $sz.ЧастотаФормирования ELSE 4 END END as Frequence 
--	, $sz.ЧислаМесяца as DaysOfMonth
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.ЧислаМесяца ELSE '' END  as DaysOfMonth
   , $st.ИмяВКонфигураторе as Сотрудник

   ,CONVERT(varchar(10),$sz.ДатаНачала,102) as BeginDate
   , $sz.EMail as EMail
   , $sz.АвтоОтправка as AutoSend
   , $sz.ВремяОтправки as Time
   
   , ISNULL(stk.DESCR,'') as Comments	
   , $sz.Дни as Дни
FROM $Справочник.ГрафикЗаказовАВМ as sz (NOLOCK)
INNER JOIN $Справочник.Контрагенты as k (NOLOCK) ON k.ID=$sz.Поставщик
INNER JOIN $Справочник.Сотрудники as st (NOLOCK) ON st.ID=$sz.Сотрудник
LEFT OUTER JOIN $Справочник.Сотрудники as stk (NOLOCK) ON stk.ID=$sz.КатегорийныйМенеджер
INNER JOIN $Справочник.ГрафикЗаказовАВМ_Магазины as sgm (NOLOCK) ON sgm.ParentExt=sz.ID
INNER JOIN $Справочник.Фирмы as f (NOLOCK) ON f.ID=$sgm.Магазин
INNER JOIN $Справочник.Супермаркеты as sp (NOLOCK) ON sp.ID=$f.Супермаркет
WHERE $sp.Доступен=2 AND sz.ISMARK=0
ORDER BY sz.DESCR, $f.КраткоеНазвание
   `
}

func GetScheduleNew() string {
	return `
SELECT
	sz.CODE
   , f.Code as FirmaID, $f.КраткоеНазвание as fname
   , k.CODE as SupplierCode, k.DESCR as Поставщик
   , $sz.РазделительЗаказов as OrderSplitMark
   , CASE WHEN $sz.Активный=0 THEN 0 ELSE $sgm.Активный END as  Активный
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.Периодичность ELSE 7 END as Периодичность  
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Понедельник > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Monday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Вторник > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as TuesDay
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Среда > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Wednesday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Четверг > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Thursday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Пятница > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Friday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Суббота > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Saturday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN CAST($sgm.ПлечоПоставки as varchar) ELSE CASE WHEN $sz.Воскресенье > 0 THEN CAST($sgm.ПлечоПоставки as varchar) ELSE '' END END as Sunday
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.ЧастотаФормирования ELSE CASE WHEN $sz.Периодичность=7 THEN $sz.ЧастотаФормирования ELSE 4 END END as Frequence 
--	, $sz.ЧислаМесяца as DaysOfMonth
   , CASE WHEN $sz.Периодичность = 30 AND SUBSTRING($sz.Дни,1,1) <> '2' THEN $sz.ЧислаМесяца ELSE '' END  as DaysOfMonth
   , $st.ИмяВКонфигураторе as Сотрудник

   ,REPLACE(CONVERT(varchar(10),$sz.ДатаНачала,102),'.','-') as BeginDate
   , $sz.EMail as EMail
   , $sz.АвтоОтправка as AutoSend
   , $sz.ВремяОтправки as Time
   
   , ISNULL(stk.DESCR,'') as Comments	
   , $sz.Дни as Дни
FROM $Справочник.ГрафикЗаказовАВМ as sz (NOLOCK)
INNER JOIN $Справочник.Контрагенты as k (NOLOCK) ON k.ID=$sz.Поставщик
INNER JOIN $Справочник.Сотрудники as st (NOLOCK) ON st.ID=$sz.Сотрудник
LEFT OUTER JOIN $Справочник.Сотрудники as stk (NOLOCK) ON stk.ID=$sz.КатегорийныйМенеджер
INNER JOIN $Справочник.ГрафикЗаказовАВМ_Магазины as sgm (NOLOCK) ON sgm.ParentExt=sz.ID
INNER JOIN $Справочник.Фирмы as f (NOLOCK) ON f.ID=$sgm.Магазин
INNER JOIN $Справочник.Супермаркеты as sp (NOLOCK) ON sp.ID=$f.Супермаркет
WHERE $sp.Доступен=2 AND sz.ISMARK=0
ORDER BY sz.DESCR, $f.КраткоеНазвание
   `
}

func Get() string {
	return `
	`
}

func InitScript() {
	ScriptMain = make(map[string]string)
	ScriptMain["ФильтрИсключений"] = GetFilterExceptions()
	ScriptMain["ПривязкуМенеджеров"] = GetSnapManagers()
	ScriptMain["СписокТоваровИзЗапускаемыхГрупп"] = GetListOfProductsFromLaunchedGroups()
	ScriptMain["КоличествоЗагруженныхМагазинов"] = GetNumberOfStoresDownloaded()
	ScriptMain["КоличествоНеПришедшихМагазинов"] = GetNumberOfNotArrivedStores()
	ScriptMain["СписокНеПришедшихМагазинов"] = GetGetListOfNotArrivedStores()
	ScriptMain["ФормированиеВыборкиДанных"] = GetGeneratingSampleData()
	ScriptMain["ПолучитьStores"] = GetStores()
	ScriptMain["ПолучитьSkuHeader"] = GetSkuHeader()
	ScriptMain["ПолучитьSkuHeaderNew"] = GetSkuHeaderNew()
	ScriptMain["ПолучитьBrands"] = GetBrands()
	ScriptMain["ПолучитьSuppliers"] = GetSuppliers()
	ScriptMain["ПолучитьSchedule"] = GetSchedule()
	ScriptMain["ПолучитьScheduleNew"] = GetScheduleNew()
}
