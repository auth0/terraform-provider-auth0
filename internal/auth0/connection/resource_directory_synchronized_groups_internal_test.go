package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

func groupIDsAttribute(groupIDs ...string) []interface{} {
	attribute := make([]interface{}, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		attribute = append(attribute, groupID)
	}

	return attribute
}

// groupsAttribute renders one block per group carrying only `id`, all a practitioner can set.
func groupsAttribute(groupIDs ...string) []interface{} {
	attribute := make([]interface{}, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		attribute = append(attribute, map[string]interface{}{"id": groupID})
	}

	return attribute
}

// resourceDataForUpdate builds the *schema.ResourceData an update sees, with the diff computed by
// the plugin SDK rather than hand-written. That is what makes these tests meaningful for a migration,
// where the question is what `GetChange` reports across two different attributes.
func resourceDataForUpdate(t *testing.T, priorAttributes map[string]interface{}, configuration map[string]interface{}) *schema.ResourceData {
	t.Helper()

	resource := NewDirectorySynchronizedGroupsResource()
	schemaMap := schema.InternalMap(resource.Schema)

	priorData := schema.TestResourceDataRaw(t, resource.Schema, priorAttributes)
	priorData.SetId("con_directorySynchronizedGroups")
	priorState := priorData.State()

	diff, err := schemaMap.Diff(
		context.Background(),
		priorState,
		terraform.NewResourceConfigRaw(configuration),
		nil,
		nil,
		true,
	)
	require.NoError(t, err)

	data, err := schemaMap.Data(priorState, diff)
	require.NoError(t, err)
	data.SetId("con_directorySynchronizedGroups")

	return data
}

// TestDirectorySynchronizedGroupsMigration drives the real plugin-SDK diff for every way a
// practitioner can move between `group_ids` and `groups`. The property being verified is that the
// migration itself is free: moving the same IDs across sends nothing, and a migration that also
// changes the IDs sends only that change.
func TestDirectorySynchronizedGroupsMigration(t *testing.T) {
	testCases := []struct {
		name                 string
		priorAttributes      map[string]interface{}
		configuration        map[string]interface{}
		expectedGroupsToAdd  []string
		expectedGroupsToRemn []string
	}{
		// The cases called out on the card, moving from the deprecated attribute to the new one.
		{
			name:                 "group_ids (1,2,3) to groups (1,2,3) synchronizes nothing",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3")},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: nil,
		},
		{
			name:                 "group_ids (1,2,3) to groups (1,2,3,4) adds only the new group",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3", "group4")},
			expectedGroupsToAdd:  []string{"group4"},
			expectedGroupsToRemn: nil,
		},
		{
			name:                 "group_ids (1,2,3) to groups (1,2) removes only the dropped group",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: []string{"group3"},
		},
		{
			name:                 "group_ids (1,2,3) to groups (4,5) replaces the whole set",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group4", "group5")},
			expectedGroupsToAdd:  []string{"group4", "group5"},
			expectedGroupsToRemn: []string{"group1", "group2", "group3"},
		},
		{
			name:                 "group_ids (1,2,3) to no groups at all unsynchronizes everything",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: []string{"group1", "group2", "group3"},
		},
		{
			name:                 "no group_ids to groups (4,5) adds both",
			priorAttributes:      map[string]interface{}{},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group4", "group5")},
			expectedGroupsToAdd:  []string{"group4", "group5"},
			expectedGroupsToRemn: nil,
		},
		{
			name:                 "empty group_ids to groups (4,5) adds both",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute()},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group4", "group5")},
			expectedGroupsToAdd:  []string{"group4", "group5"},
			expectedGroupsToRemn: nil,
		},

		// The same moves in reverse: nothing forces a practitioner to migrate, so going back has
		// to behave symmetrically.
		{
			name:                 "groups (1,2,3) to group_ids (1,2,3) synchronizes nothing",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: nil,
		},
		{
			name:                 "groups (1,2,3) to group_ids (4,5) replaces the whole set",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{"group_ids": groupIDsAttribute("group4", "group5")},
			expectedGroupsToAdd:  []string{"group4", "group5"},
			expectedGroupsToRemn: []string{"group1", "group2", "group3"},
		},
		{
			name:                 "groups (1,2,3) to no group_ids at all unsynchronizes everything",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3")},
			configuration:        map[string]interface{}{},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: []string{"group1", "group2", "group3"},
		},

		// Changes that stay on one attribute, which is what most practitioners keep doing.
		{
			name:                 "group_ids (1,2) to group_ids (2,3) adds and removes one each",
			priorAttributes:      map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2")},
			configuration:        map[string]interface{}{"group_ids": groupIDsAttribute("group2", "group3")},
			expectedGroupsToAdd:  []string{"group3"},
			expectedGroupsToRemn: []string{"group1"},
		},
		{
			name:                 "groups (1,2) to groups (2,3) adds and removes one each",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group2", "group3")},
			expectedGroupsToAdd:  []string{"group3"},
			expectedGroupsToRemn: []string{"group1"},
		},
		{
			name:                 "groups (1,2) unchanged synchronizes nothing",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			configuration:        map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: nil,
		},

		// An import lands the IDs in `groups`, so a practitioner importing into a `group_ids`
		// configuration takes this path on their first apply. It has to be free.
		{
			name:                 "imported into groups, configured as group_ids, synchronizes nothing",
			priorAttributes:      map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			configuration:        map[string]interface{}{"group_ids": groupIDsAttribute("group2", "group1")},
			expectedGroupsToAdd:  nil,
			expectedGroupsToRemn: nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.priorAttributes["connection_id"] = "con_directorySynchronizedGroups"
			testCase.configuration["connection_id"] = "con_directorySynchronizedGroups"

			data := resourceDataForUpdate(t, testCase.priorAttributes, testCase.configuration)

			groupIDsToAdd, groupIDsToRemove := diffGroupIDs(groupIDsChange(data))

			assert.ElementsMatch(t, testCase.expectedGroupsToAdd, groupIDsToAdd)
			assert.ElementsMatch(t, testCase.expectedGroupsToRemn, groupIDsToRemove)
		})
	}
}

// TestUpdateDirectorySynchronizedGroupsSendsOnlyTheDifference takes the same migration cases down
// to the wire, because an empty difference only matters if it also means no call is made.
func TestUpdateDirectorySynchronizedGroupsSendsOnlyTheDifference(t *testing.T) {
	testCases := []struct {
		name             string
		priorAttributes  map[string]interface{}
		configuration    map[string]interface{}
		expectedRequests []expectedWrite
	}{
		{
			name:             "moving the same group IDs from group_ids to groups sends nothing",
			priorAttributes:  map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:    map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3")},
			expectedRequests: nil,
		},
		{
			name:            "migrating and adding a group sends one add",
			priorAttributes: map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:   map[string]interface{}{"groups": groupsAttribute("group1", "group2", "group3", "group4")},
			expectedRequests: []expectedWrite{
				{Method: http.MethodPost, GroupIDs: []string{"group4"}},
			},
		},
		{
			name:            "migrating and dropping a group sends one remove",
			priorAttributes: map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:   map[string]interface{}{"groups": groupsAttribute("group1", "group2")},
			expectedRequests: []expectedWrite{
				{Method: http.MethodDelete, GroupIDs: []string{"group3"}},
			},
		},
		{
			name:            "migrating to a disjoint set removes before it adds",
			priorAttributes: map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:   map[string]interface{}{"groups": groupsAttribute("group4", "group5")},
			expectedRequests: []expectedWrite{
				{Method: http.MethodDelete, GroupIDs: []string{"group1", "group2", "group3"}},
				{Method: http.MethodPost, GroupIDs: []string{"group4", "group5"}},
			},
		},
		{
			name:            "dropping the attribute entirely removes every group",
			priorAttributes: map[string]interface{}{"group_ids": groupIDsAttribute("group1", "group2", "group3")},
			configuration:   map[string]interface{}{},
			expectedRequests: []expectedWrite{
				{Method: http.MethodDelete, GroupIDs: []string{"group1", "group2", "group3"}},
			},
		},
		{
			name:            "adopting groups from nothing adds every group",
			priorAttributes: map[string]interface{}{},
			configuration:   map[string]interface{}{"groups": groupsAttribute("group4", "group5")},
			expectedRequests: []expectedWrite{
				{Method: http.MethodPost, GroupIDs: []string{"group4", "group5"}},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.priorAttributes["connection_id"] = "con_directorySynchronizedGroups"
			testCase.configuration["connection_id"] = "con_directorySynchronizedGroups"

			data := resourceDataForUpdate(t, testCase.priorAttributes, testCase.configuration)
			api, requests := newRecordingAPI(t, `{"groups":[]}`)

			diagnostics := updateDirectorySynchronizedGroups(context.Background(), data, api)

			assert.False(t, diagnostics.HasError())
			assertWrites(t, testCase.expectedRequests, requests)
		})
	}
}

// TestUpdateDirectorySynchronizedGroupsRemovesGroupIDsOnly guards the remove payload shape: the
// endpoint answers anything richer than an ID with `400 Additional properties not allowed: name`.
func TestUpdateDirectorySynchronizedGroupsRemovesGroupIDsOnly(t *testing.T) {
	data := resourceDataForUpdate(t,
		map[string]interface{}{
			"connection_id": "con_directorySynchronizedGroups",
			"groups": []interface{}{
				map[string]interface{}{"id": "group1", "name": "Engineering", "email": "engineering@example.com", "direct_members_count": 7},
			},
		},
		map[string]interface{}{"connection_id": "con_directorySynchronizedGroups"},
	)

	api, requests := newRecordingAPI(t, `{"groups":[]}`)

	diagnostics := updateDirectorySynchronizedGroups(context.Background(), data, api)

	assert.False(t, diagnostics.HasError())
	require.Len(t, requests.writes(), 1)
	assert.Equal(t, http.MethodDelete, requests.writes()[0].Method)
	assert.Equal(t,
		[]map[string]interface{}{{"id": "group1"}},
		requests.writes()[0].Groups,
		"the remove payload must carry the group ID alone, with none of the metadata state holds",
	)
}

// TestCreateDirectorySynchronizedGroupsReplacesTheWholeSet covers create replacing rather than
// diffing, since the connection may already have groups synchronized outside Terraform.
func TestCreateDirectorySynchronizedGroupsReplacesTheWholeSet(t *testing.T) {
	t.Run("it replaces with the configured groups", func(t *testing.T) {
		data := resourceDataForUpdate(t, nil, map[string]interface{}{
			"connection_id": "con_directorySynchronizedGroups",
			"groups":        groupsAttribute("group1", "group2"),
		})
		api, requests := newRecordingAPI(t, `{"groups":[{"id":"group1"},{"id":"group2"}]}`)

		diagnostics := createDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assertWrites(t, []expectedWrite{
			{Method: http.MethodPut, GroupIDs: []string{"group1", "group2"}},
		}, requests)
		assert.Equal(t, "con_directorySynchronizedGroups", data.Id())
	})

	t.Run("it replaces with an empty set when no groups are configured", func(t *testing.T) {
		data := resourceDataForUpdate(t, nil, map[string]interface{}{
			"connection_id": "con_directorySynchronizedGroups",
		})
		api, requests := newRecordingAPI(t, `{"groups":[]}`)

		diagnostics := createDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assertWrites(t, []expectedWrite{
			{Method: http.MethodPut, GroupIDs: []string{}},
		}, requests)
	})
}

// TestDeleteDirectorySynchronizedGroups covers delete unsynchronizing every group with an empty
// replace, rather than deleting the connection's configuration.
func TestDeleteDirectorySynchronizedGroups(t *testing.T) {
	data := resourceDataForUpdate(t, map[string]interface{}{
		"connection_id": "con_directorySynchronizedGroups",
		"groups":        groupsAttribute("group1", "group2"),
	}, map[string]interface{}{
		"connection_id": "con_directorySynchronizedGroups",
		"groups":        groupsAttribute("group1", "group2"),
	})
	api, requests := newRecordingAPI(t, `{"groups":[]}`)

	diagnostics := deleteDirectorySynchronizedGroups(context.Background(), data, api)

	assert.False(t, diagnostics.HasError())
	assertWrites(t, []expectedWrite{
		{Method: http.MethodPut, GroupIDs: []string{}},
	}, requests)
}

// TestReadDirectorySynchronizedGroupsWritesOneAttribute pins the read to whichever attribute state
// already uses, since populating the unused one would propose emptying it on every later plan.
func TestReadDirectorySynchronizedGroupsWritesOneAttribute(t *testing.T) {
	const apiResponse = `{"groups":[
		{"id":"group1","name":"Engineering","email":"engineering@example.com","direct_members_count":7},
		{"id":"group2"}
	]}`

	t.Run("it writes group_ids when state already uses group_ids", func(t *testing.T) {
		data := resourceDataForUpdate(t,
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "group_ids": groupIDsAttribute("group1")},
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "group_ids": groupIDsAttribute("group1")},
		)
		api, _ := newRecordingAPI(t, apiResponse)

		diagnostics := readDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assert.ElementsMatch(t, []interface{}{"group1", "group2"}, data.Get("group_ids").(*schema.Set).List())
		assert.Empty(t, data.Get("groups").(*schema.Set).List())
	})

	t.Run("it writes groups when state already uses groups", func(t *testing.T) {
		data := resourceDataForUpdate(t,
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "groups": groupsAttribute("group1")},
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "groups": groupsAttribute("group1")},
		)
		api, _ := newRecordingAPI(t, apiResponse)

		diagnostics := readDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assert.Empty(t, data.Get("group_ids").(*schema.Set).List())
		assert.ElementsMatch(t, []interface{}{
			map[string]interface{}{"id": "group1", "name": "Engineering", "email": "engineering@example.com", "direct_members_count": 7},
			map[string]interface{}{"id": "group2", "name": "", "email": "", "direct_members_count": 0},
		}, data.Get("groups").(*schema.Set).List())
	})

	t.Run("it writes groups on an import, where neither attribute is in state", func(t *testing.T) {
		data := NewDirectorySynchronizedGroupsResource().Data(nil)
		data.SetId("con_directorySynchronizedGroups")
		api, _ := newRecordingAPI(t, apiResponse)

		diagnostics := readDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assert.Equal(t, "con_directorySynchronizedGroups", data.Get("connection_id"))
		assert.Empty(t, data.Get("group_ids").(*schema.Set).List())
		assert.Len(t, data.Get("groups").(*schema.Set).List(), 2)
	})

	t.Run("it empties group_ids when every group was unsynchronized elsewhere", func(t *testing.T) {
		data := resourceDataForUpdate(t,
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "group_ids": groupIDsAttribute("group1")},
			map[string]interface{}{"connection_id": "con_directorySynchronizedGroups", "group_ids": groupIDsAttribute("group1")},
		)
		api, _ := newRecordingAPI(t, `{"groups":[]}`)

		diagnostics := readDirectorySynchronizedGroups(context.Background(), data, api)

		assert.False(t, diagnostics.HasError())
		assert.Empty(t, data.Get("group_ids").(*schema.Set).List())
		assert.Empty(t, data.Get("groups").(*schema.Set).List())
	})
}

// TestReadDirectorySynchronizedGroupsWhenConnectionIsGone covers a connection deleted outside
// Terraform: the read drops the resource from state with a warning rather than wedging the plan.
func TestReadDirectorySynchronizedGroupsWhenConnectionIsGone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"statusCode":404,"error":"Not Found","message":"The connection does not exist."}`))
	}))
	t.Cleanup(server.Close)

	data := NewDirectorySynchronizedGroupsResource().Data(nil)
	data.SetId("con_directorySynchronizedGroups")

	diagnostics := readDirectorySynchronizedGroups(context.Background(), data, newAPIAt(server.URL))

	assert.False(t, diagnostics.HasError())
	require.Len(t, diagnostics, 1)
	assert.Equal(t, "Resource not found, removed from state", diagnostics[0].Summary)
	assert.Empty(t, data.Id())
}

// TestReadDirectorySynchronizedGroupsDataSourceWritesBothAttributes covers the data source writing
// both attributes: they are outputs, so writing both cannot produce a diff.
func TestReadDirectorySynchronizedGroupsDataSourceWritesBothAttributes(t *testing.T) {
	data := schema.TestResourceDataRaw(t, getDirectorySynchronizedGroupsDataSourceSchema(), map[string]interface{}{
		"connection_id": "con_directorySynchronizedGroups",
		"query":         "name:engineering*",
	})
	api, requests := newRecordingAPI(t, `{"groups":[{"id":"group1","name":"Engineering","email":"engineering@example.com","direct_members_count":7}]}`)

	diagnostics := readDirectorySynchronizedGroupsDataSource(context.Background(), data, api)

	assert.False(t, diagnostics.HasError())
	assert.Equal(t, "con_directorySynchronizedGroups", data.Id())
	assert.Equal(t, []interface{}{"group1"}, data.Get("group_ids").(*schema.Set).List())
	assert.Equal(t, []interface{}{
		map[string]interface{}{"id": "group1", "name": "Engineering", "email": "engineering@example.com", "direct_members_count": 7},
	}, data.Get("groups").(*schema.Set).List())

	// The query reaches the API verbatim, rather than being validated or rewritten here.
	assert.Equal(t, []string{"name:engineering*"}, requests.queries())
}

// TestGetAllGroupsFollowsPagination guards against a partial read: the resource manages the whole
// set, so stopping at the first page would look like the later pages were unsynchronized elsewhere.
func TestGetAllGroupsFollowsPagination(t *testing.T) {
	pages := []string{
		`{"groups":[{"id":"group1"}],"next":"cursorToSecondPage"}`,
		`{"groups":[{"id":"group2"}],"next":"cursorToThirdPage"}`,
		`{"groups":[{"id":"group3"}]}`,
	}

	var cursors []string
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cursors = append(cursors, request.URL.Query().Get("from"))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(pages[page]))
		page++
	}))
	t.Cleanup(server.Close)

	groups, err := getAllGroups(context.Background(), newAPIAt(server.URL).GetAPIV3(), "con_directorySynchronizedGroups", "")

	assert.NoError(t, err)
	assert.Equal(t, []string{"group1", "group2", "group3"}, flattenGroupIDs(groups))
	assert.Equal(t, []string{"", "cursorToSecondPage", "cursorToThirdPage"}, cursors)
}

// TestChunkGroupIDs guards the boundary the API enforces inclusively: 100 groups per add or
// remove call is accepted and 101 is rejected with `Array is too long (101), maximum 100`.
func TestChunkGroupIDs(t *testing.T) {
	testCases := []struct {
		groupIDCount           int
		expectedChunkSizes     []int
		expectedChunkCountText string
	}{
		{groupIDCount: 0, expectedChunkSizes: nil},
		{groupIDCount: 1, expectedChunkSizes: []int{1}},
		{groupIDCount: 99, expectedChunkSizes: []int{99}},
		{groupIDCount: 100, expectedChunkSizes: []int{100}},
		{groupIDCount: 101, expectedChunkSizes: []int{100, 1}},
		{groupIDCount: 200, expectedChunkSizes: []int{100, 100}},
		{groupIDCount: 201, expectedChunkSizes: []int{100, 100, 1}},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(fmt.Sprintf("%d group IDs", testCase.groupIDCount), func(t *testing.T) {
			groupIDs := generateGroupIDs(testCase.groupIDCount)

			chunks := chunkGroupIDs(groupIDs)

			chunkSizes := make([]int, 0, len(chunks))
			var chunked []string
			for _, chunk := range chunks {
				chunkSizes = append(chunkSizes, len(chunk))
				chunked = append(chunked, chunk...)
			}

			assert.Equal(t, testCase.expectedChunkSizes, nilIfEmpty(chunkSizes))
			assert.Equal(t, groupIDs, chunked, "chunking must preserve every group ID exactly once, in order")
		})
	}
}

// TestAddAndRemoveGroupsChunkAtTheAPILimit takes the boundary to the wire: 101 groups leave as two
// calls of 100 and 1, not as one call the API rejects.
func TestAddAndRemoveGroupsChunkAtTheAPILimit(t *testing.T) {
	groupIDs := generateGroupIDs(101)

	t.Run("add", func(t *testing.T) {
		api, requests := newRecordingAPI(t, `{"groups":[]}`)

		assert.NoError(t, addGroups(context.Background(), api.GetAPIV3(), "con_directorySynchronizedGroups", groupIDs))

		writes := requests.writes()
		require.Len(t, writes, 2)
		assert.Equal(t, groupIDs[:100], writes[0].GroupIDs)
		assert.Equal(t, groupIDs[100:], writes[1].GroupIDs)
	})

	t.Run("remove", func(t *testing.T) {
		api, requests := newRecordingAPI(t, `{"groups":[]}`)

		assert.NoError(t, removeGroups(context.Background(), api.GetAPIV3(), "con_directorySynchronizedGroups", groupIDs))

		writes := requests.writes()
		require.Len(t, writes, 2)
		assert.Equal(t, groupIDs[:100], writes[0].GroupIDs)
		assert.Equal(t, groupIDs[100:], writes[1].GroupIDs)
	})

	t.Run("neither issues a call when there is nothing to do", func(t *testing.T) {
		api, requests := newRecordingAPI(t, `{"groups":[]}`)

		assert.NoError(t, addGroups(context.Background(), api.GetAPIV3(), "con_directorySynchronizedGroups", nil))
		assert.NoError(t, removeGroups(context.Background(), api.GetAPIV3(), "con_directorySynchronizedGroups", nil))

		assert.Empty(t, requests.writes())
	})
}

// TestPutGroupsIsNotChunked records that replace is sent whole. Its limit is the documented 10000
// ceiling, not the 100 add and remove impose, and chunking it would leave only the last chunk.
func TestPutGroupsIsNotChunked(t *testing.T) {
	groupIDs := generateGroupIDs(101)
	api, requests := newRecordingAPI(t, `{"groups":[]}`)

	assert.NoError(t, putGroups(context.Background(), api.GetAPIV3(), "con_directorySynchronizedGroups", groupIDs))

	writes := requests.writes()
	require.Len(t, writes, 1)
	assert.Equal(t, http.MethodPut, writes[0].Method)
	assert.Equal(t, groupIDs, writes[0].GroupIDs)
}

func TestDiffGroupIDs(t *testing.T) {
	testCases := []struct {
		name                    string
		prior                   []string
		desired                 []string
		expectedGroupIDsToAdd   []string
		expectedGroupIDsToRemov []string
	}{
		{
			name:    "both empty",
			prior:   nil,
			desired: nil,
		},
		{
			name:                  "everything desired is added when nothing is synchronized",
			prior:                 nil,
			desired:               []string{"group1", "group2"},
			expectedGroupIDsToAdd: []string{"group1", "group2"},
		},
		{
			name:                    "everything synchronized is removed when nothing is desired",
			prior:                   []string{"group1", "group2"},
			desired:                 nil,
			expectedGroupIDsToRemov: []string{"group1", "group2"},
		},
		{
			name:    "identical sets need no calls",
			prior:   []string{"group1", "group2"},
			desired: []string{"group1", "group2"},
		},
		{
			name:    "reordering is not a change",
			prior:   []string{"group1", "group2", "group3"},
			desired: []string{"group3", "group1", "group2"},
		},
		{
			name:                    "overlapping sets add and remove only the difference",
			prior:                   []string{"group1", "group2", "group3"},
			desired:                 []string{"group2", "group4"},
			expectedGroupIDsToAdd:   []string{"group4"},
			expectedGroupIDsToRemov: []string{"group1", "group3"},
		},
		{
			// The attributes are sets, so Terraform cannot hand us a duplicate. The guard is for a
			// future caller assembling the slices itself.
			name:                  "a duplicate in the desired set is added once",
			prior:                 nil,
			desired:               []string{"group1", "group1"},
			expectedGroupIDsToAdd: []string{"group1", "group1"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			groupIDsToAdd, groupIDsToRemove := diffGroupIDs(testCase.prior, testCase.desired)

			assert.Equal(t, testCase.expectedGroupIDsToAdd, groupIDsToAdd)
			assert.Equal(t, testCase.expectedGroupIDsToRemov, groupIDsToRemove)
		})
	}
}

func TestMergeGroupIDs(t *testing.T) {
	resourceSchema := NewDirectorySynchronizedGroupsResource().Schema

	newGroupIDSet := func(groupIDs ...string) *schema.Set {
		return schema.NewSet(schema.HashString, groupIDsAttribute(groupIDs...))
	}
	newGroupSet := func(groups ...interface{}) *schema.Set {
		return schema.NewSet(schema.HashResource(resourceSchema["groups"].Elem.(*schema.Resource)), groups)
	}

	t.Run("it collects from group_ids alone", func(t *testing.T) {
		assert.ElementsMatch(t,
			[]string{"group1", "group2"},
			mergeGroupIDs(newGroupIDSet("group1", "group2"), newGroupSet()),
		)
	})

	t.Run("it collects from groups alone", func(t *testing.T) {
		assert.ElementsMatch(t,
			[]string{"group1", "group2"},
			mergeGroupIDs(newGroupIDSet(), newGroupSet(groupsAttribute("group1", "group2")...)),
		)
	})

	t.Run("it collects from both, which is what makes a migration send nothing", func(t *testing.T) {
		assert.ElementsMatch(t,
			[]string{"group1", "group2"},
			mergeGroupIDs(newGroupIDSet("group1"), newGroupSet(groupsAttribute("group2")...)),
		)
	})

	t.Run("it returns an empty slice rather than nil when there is nothing to collect", func(t *testing.T) {
		assert.Equal(t, []string{}, mergeGroupIDs(newGroupIDSet(), newGroupSet()))
		assert.Equal(t, []string{}, mergeGroupIDs(nil, nil))
	})

	t.Run("it skips empty group IDs", func(t *testing.T) {
		assert.Equal(t,
			[]string{"group1"},
			mergeGroupIDs(newGroupIDSet("group1", ""), newGroupSet(map[string]interface{}{"id": ""})),
		)
	})
}

func TestFlattenGroups(t *testing.T) {
	t.Run("it carries every field the API populated", func(t *testing.T) {
		assert.Equal(t,
			[]interface{}{map[string]interface{}{
				"id":                   "group1",
				"name":                 "Engineering",
				"email":                "engineering@example.com",
				"direct_members_count": 7,
			}},
			flattenGroups([]*managementv3.SynchronizedGroupPayload{{
				ID:                 "group1",
				Name:               pointerTo("Engineering"),
				Email:              pointerTo("engineering@example.com"),
				DirectMembersCount: pointerTo(7),
			}}),
		)
	})

	t.Run("it reads absent metadata as the zero value", func(t *testing.T) {
		// The API omits the metadata when it has not been populated. The zero value is how that
		// should read, and it keeps the plan empty.
		assert.Equal(t,
			[]interface{}{map[string]interface{}{
				"id":                   "group1",
				"name":                 "",
				"email":                "",
				"direct_members_count": 0,
			}},
			flattenGroups([]*managementv3.SynchronizedGroupPayload{{ID: "group1"}}),
		)
	})

	t.Run("it returns an empty slice rather than nil for no groups", func(t *testing.T) {
		assert.Equal(t, []interface{}{}, flattenGroups(nil))
	})
}

func TestFlattenGroupIDs(t *testing.T) {
	assert.Equal(t,
		[]string{"group1", "group2"},
		flattenGroupIDs([]*managementv3.SynchronizedGroupPayload{
			{ID: "group1", Name: pointerTo("Engineering")},
			{ID: "group2"},
		}),
	)
	assert.Equal(t, []string{}, flattenGroupIDs(nil))
}

// TestDirectorySynchronizedGroupsSchemaConflict records how a practitioner has to write the
// migration: `ConflictsWith` fires on an attribute being present at all, not on it holding anything,
// so `group_ids = []` beside a `groups` block is an error. The line has to be deleted, not emptied.
// Setting neither stays valid.
func TestDirectorySynchronizedGroupsSchemaConflict(t *testing.T) {
	schemaMap := schema.InternalMap(NewDirectorySynchronizedGroupsResource().Schema)

	validate := func(configuration map[string]interface{}) []string {
		configuration["connection_id"] = "con_directorySynchronizedGroups"

		var summaries []string
		for _, diagnostic := range schemaMap.Validate(terraform.NewResourceConfigRaw(configuration)) {
			if diagnostic.Severity == 0 {
				summaries = append(summaries, diagnostic.Summary)
			}
		}

		return summaries
	}

	t.Run("group_ids alone is valid, and warns that it is deprecated", func(t *testing.T) {
		assert.Empty(t, validate(map[string]interface{}{"group_ids": groupIDsAttribute("group1")}))

		diagnostics := schemaMap.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
			"connection_id": "con_directorySynchronizedGroups",
			"group_ids":     groupIDsAttribute("group1"),
		}))
		require.Len(t, diagnostics, 1)
		assert.Equal(t, "Argument is deprecated", diagnostics[0].Summary)
	})

	t.Run("groups alone is valid and warns about nothing", func(t *testing.T) {
		assert.Empty(t, schemaMap.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
			"connection_id": "con_directorySynchronizedGroups",
			"groups":        groupsAttribute("group1"),
		})))
	})

	t.Run("neither is valid, because that is how every group is unsynchronized", func(t *testing.T) {
		assert.Empty(t, validate(map[string]interface{}{}))
	})

	t.Run("both together conflict", func(t *testing.T) {
		assert.Equal(t,
			[]string{"Conflicting configuration arguments", "Conflicting configuration arguments"},
			validate(map[string]interface{}{
				"group_ids": groupIDsAttribute("group1"),
				"groups":    groupsAttribute("group1"),
			}),
		)
	})

	t.Run("an emptied group_ids still conflicts, so a migration has to delete the line", func(t *testing.T) {
		assert.Equal(t,
			[]string{"Conflicting configuration arguments", "Conflicting configuration arguments"},
			validate(map[string]interface{}{
				"group_ids": groupIDsAttribute(),
				"groups":    groupsAttribute("group1"),
			}),
		)
	})
}

type recordedRequest struct {
	Method   string
	GroupIDs []string

	// Groups holds the payload objects verbatim, for the assertions on which properties a payload
	// carried rather than only which groups it named.
	Groups []map[string]interface{}
}

type expectedWrite struct {
	Method   string
	GroupIDs []string
}

type recordedRequests struct {
	requests []recordedRequest
	query    []string
}

// writes drops the reads, which every mutating call is followed by and which say nothing about the
// difference the update computed.
func (recorded *recordedRequests) writes() []recordedRequest {
	var writes []recordedRequest
	for _, request := range recorded.requests {
		if request.Method != http.MethodGet {
			writes = append(writes, request)
		}
	}

	return writes
}

func (recorded *recordedRequests) queries() []string {
	return recorded.query
}

// assertWrites compares methods in order, because removing before adding is deliberate, but the IDs
// within each call without regard to order, because they come from a set.
func assertWrites(t *testing.T, expected []expectedWrite, recorded *recordedRequests) {
	t.Helper()

	writes := recorded.writes()
	require.Len(t, writes, len(expected))

	for i, expectedWrite := range expected {
		assert.Equal(t, expectedWrite.Method, writes[i].Method)
		assert.ElementsMatch(t, expectedWrite.GroupIDs, writes[i].GroupIDs)
	}
}

// newRecordingAPI points a *config.Config at a server that records every request and answers each
// read with the given JSON.
func newRecordingAPI(t *testing.T, listResponse string) (*config.Config, *recordedRequests) {
	t.Helper()

	recorded := &recordedRequests{}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)

		recorded.requests = append(recorded.requests, recordedRequest{
			Method:   request.Method,
			GroupIDs: groupIDsInPayload(t, body),
			Groups:   groupsInPayload(t, body),
		})

		if request.Method == http.MethodGet {
			recorded.query = append(recorded.query, request.URL.Query().Get("q"))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(listResponse))

			return
		}

		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	return newAPIAt(server.URL), recorded
}

func newAPIAt(baseURL string) *config.Config {
	return config.NewWithV3(nil, managementv3client.NewWithOptions(
		option.WithBaseURL(baseURL),
		option.WithToken("aTestToken"),
		option.WithoutRetries(),
	))
}

func groupsInPayload(t *testing.T, body []byte) []map[string]interface{} {
	t.Helper()

	if len(body) == 0 {
		return nil
	}

	var payload struct {
		Groups []map[string]interface{} `json:"groups"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	return payload.Groups
}

func groupIDsInPayload(t *testing.T, body []byte) []string {
	t.Helper()

	if len(body) == 0 {
		return nil
	}

	groupIDs := make([]string, 0)
	for _, group := range groupsInPayload(t, body) {
		groupID, ok := group["id"].(string)
		require.True(t, ok, "every group in a payload must carry an id")
		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs
}

func generateGroupIDs(count int) []string {
	if count == 0 {
		return nil
	}

	groupIDs := make([]string, 0, count)
	for i := range count {
		groupIDs = append(groupIDs, "group"+strconv.Itoa(i))
	}

	return groupIDs
}

func nilIfEmpty(values []int) []int {
	if len(values) == 0 {
		return nil
	}

	return values
}

func pointerTo[T any](value T) *T {
	return &value
}
