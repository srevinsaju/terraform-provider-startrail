package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

//	{
//	 "access": [
//	   {
//	     "auth": true,
//	     "endpoint": "string",
//	     "internal": true
//	   }
//	 ],
//	 "description": "This is a hello world service.",
//	 "disabled": true,
//	 "environment": "production",
//	 "logging": {
//	   "additionalProp1": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   },
//	   "additionalProp2": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   },
//	   "additionalProp3": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   }
//	 },
//	 "metadata": {
//	   "labels": {
//	     "additionalProp1": "string",
//	     "additionalProp2": "string",
//	     "additionalProp3": "string"
//	   }
//	 },
//	 "name": "hello-world",
//	 "remarks": "Make sure this service prints hello world on /",
//	 "sources": {
//	   "additionalProp1": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   },
//	   "additionalProp2": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   },
//	   "additionalProp3": {
//	     "labels": {
//	       "additionalProp1": "string",
//	       "additionalProp2": "string",
//	       "additionalProp3": "string"
//	     }
//	   }
//	 },
//	 "tenant": "startrail",
//	 "updated_at": "2021-01-01T00:00:00.000000",
//	 "updated_by": "string",
//	 "updated_date": "string"
//	}
//
// ServiceModel describes the resource data model.
type ServiceModel struct {
	Id          types.String                  `tfsdk:"id"`
	Access      []ServiceResourceModelAccess  `tfsdk:"access"`
	Description types.String                  `tfsdk:"description"`
	Disabled    types.Bool                    `tfsdk:"disabled"`
	Environment types.String                  `tfsdk:"environment"`
	Logging     []ServiceResourceModelLogging `tfsdk:"logging"`
	Metadata    *ServiceResourceModelMetadata `tfsdk:"metadata"`
	Name        types.String                  `tfsdk:"name"`
	Remarks     types.String                  `tfsdk:"remarks"`
	Sources     []ServiceResourceM0delSource  `tfsdk:"source"`
}
